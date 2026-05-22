package daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/copiuumgroup/entropy-cli/internal/database
)

// DownloadWorkerPool manages concurrent downloads.
type DownloadWorkerPool struct {
	workerCount int
	queue       chan *database.Download
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewDownloadWorkerPool creates a new worker pool.
func NewDownloadWorkerPool(workerCount int) *DownloadWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &DownloadWorkerPool{
		workerCount: workerCount,
		queue:       make(chan *database.Download, workerCount*2),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the worker pool.
func (p *DownloadWorkerPool) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Stop stops the worker pool gracefully.
func (p *DownloadWorkerPool) Stop() {
	close(p.queue)
	p.wg.Wait()
	p.cancel()
}

// Enqueue adds a download to the queue.
func (p *DownloadWorkerPool) Enqueue(download *database.Download) error {
	select {
	case p.queue <- download:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool is stopped")
	}
}

// worker processes downloads from the queue.
func (p *DownloadWorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case download, ok := <-p.queue:
			if !ok {
				return
			}

			if download != nil {
				p.processDownload(download)
			}

		case <-p.ctx.Done():
			return
		}
	}
}

// processDownload processes a single download.
func (p *DownloadWorkerPool) processDownload(download *database.Download) {
	download.Status = "downloading"
	download.StartedAt = now()
	database.UpdateDownload(download)

	// Simulate download process
	// In production, this would call the actual ingest engine
	for i := 0; i <= 100; i += 10 {
		select {
		case <-p.ctx.Done():
			download.Status = "cancelled"
			database.UpdateDownload(download)
			return
		default:
		}

		download.Progress = float64(i)
		download.Speed = fmt.Sprintf("%.1f MB/s", float64(i)/10.0)
		database.UpdateDownload(download)
		time.Sleep(time.Millisecond * 100)
	}

	download.Status = "done"
	download.Progress = 100
	download.CompletedAt = now()
	database.UpdateDownload(download)
}

// Daemon represents the background daemon.
type Daemon struct {
	workerPool *DownloadWorkerPool
	ticker     *time.Ticker
}

// NewDaemon creates a new daemon.
func NewDaemon(workerCount int) *Daemon {
	return &Daemon{
		workerPool: NewDownloadWorkerPool(workerCount),
	}
}

// Start starts the daemon.
func (d *Daemon) Start() error {
	d.workerPool.Start()
	d.startCacheCleaner()
	fmt.Println("Daemon started")
	return nil
}

// Stop stops the daemon gracefully.
func (d *Daemon) Stop() {
	if d.ticker != nil {
		d.ticker.Stop()
	}
	d.workerPool.Stop()
	fmt.Println("Daemon stopped")
}

// EnqueueDownload adds a download to the worker pool.
func (d *Daemon) EnqueueDownload(download *database.Download) error {
	return d.workerPool.Enqueue(download)
}

// startCacheCleaner periodically cleans expired search cache.
func (d *Daemon) startCacheCleaner() {
	d.ticker = time.NewTicker(1 * time.Hour)
	go func() {
		for range d.ticker.C {
			if err := database.CleanExpiredCache(); err != nil {
				fmt.Printf("Error cleaning cache: %v\n", err)
			}
		}
	}()
}

func now() *time.Time {
	n := time.Now()
	return &n
}
