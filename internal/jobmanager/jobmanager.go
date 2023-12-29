package jobmanager

import (
	"context"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/errors"
	"github.com/MlDenis/diploma-wannabe-v2/internal/logger"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
)

type Job struct {
	orderNumber string
	username    string
	cancel      context.CancelFunc
}

type Jobmanager struct {
	AccrualURL string
	Jobs       chan *Job
	Cursor     *db.Cursor
	mu         sync.Mutex
	client     *resty.Client
	context    context.Context
	Shutdown   context.CancelFunc
}

const JOBTIMEOUT = 10

func NewJobmanager(cursor *db.Cursor, accrualURL string, parent *context.Context) *Jobmanager {
	ctx, cancel := context.WithCancel(*parent)
	return &Jobmanager{
		AccrualURL: accrualURL,
		Jobs:       make(chan *Job),
		Cursor:     cursor,
		client:     resty.New().SetBaseURL(accrualURL),
		context:    ctx,
		Shutdown:   cancel,
	}
}

func (jm *Jobmanager) AskAccrual(url string, number string) (*models.AccrualResponse, int, error) {
	acc := models.AccrualResponse{}
	req := jm.client.R().
		SetResult(&acc).
		SetPathParam("number", number)

	resp, err := req.Get("/api/orders/{number}")
	if err != nil {
		logger.ErrorLog.Printf("Error getting order from accrual: %e", err)
		return nil, 0, err
	}
	logger.InfoLog.Printf("Accrual GET status code: %d", resp.StatusCode())
	if resp.StatusCode() == 429 {
		return nil, resp.StatusCode(), nil
	}
	if resp.StatusCode() == 204 {
		return &models.AccrualResponse{Status: "NEW"}, 204, nil
	}
	return &acc, resp.StatusCode(), nil
}

func (jm *Jobmanager) RunJob(job *Job) {
	response, statusCode, err := jm.AskAccrual(jm.AccrualURL, job.orderNumber)
	if err != nil {
		job.cancel()
	}
	if statusCode == 429 {
		time.Sleep(time.Second)
	}
	for response.Status != "INVALID" && response.Status != "PROCESSED" {
		response, statusCode, err = jm.AskAccrual(jm.AccrualURL, job.orderNumber)
		if err != nil {
			job.cancel()
		}
		if statusCode == 429 {
			time.Sleep(time.Second)
			continue
		}
		jm.mu.Lock()
		jm.Cursor.UpdateOrder(job.username, response)
		jm.mu.Unlock()
	}
	jm.mu.Lock()
	jm.Cursor.UpdateOrder(job.username, response)
	jm.Cursor.UpdateUserBalance(job.username, &models.Balance{
		Current:   response.Accrual,
		Withdrawn: 0.0,
	})
	jm.mu.Unlock()
	logger.InfoLog.Println("Job finished")
}

func (jm *Jobmanager) AddJob(orderNumber string, username string) error {
	_, cancel := context.WithTimeout(jm.context, JOBTIMEOUT*time.Second)
	jm.Jobs <- &Job{orderNumber: orderNumber, username: username, cancel: cancel}
	if jm.Jobs == nil {
		cancel()
		return errors.ErrJobChannelClosed
	}
	return nil
}

func (jm *Jobmanager) ManageJobs(accrualURL string) {
	var wg sync.WaitGroup
	select {
	case <-jm.context.Done():
		close(jm.Jobs)
	default:
		for job := range jm.Jobs {
			wg.Add(1)
			logger.InfoLog.Printf("Running job for order %s", job.orderNumber)
			go jm.RunJob(job)
			wg.Done()
		}
	}
	wg.Wait()
}
