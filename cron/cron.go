package cron

import (
	"context"
	"log"
	"policy-backend/search"
	"time"

	"gorm.io/gorm"
)

// CronJob 定时任务管理器
type CronJob struct {
	db          *gorm.DB
	searchH     *search.Handler
	ctx         context.Context
	cancelFunc  context.CancelFunc
}

// NewCronJob 创建新的定时任务管理器
func NewCronJob(db *gorm.DB, searchH *search.Handler) *CronJob {
	ctx, cancel := context.WithCancel(context.Background())
	return &CronJob{
		db:         db,
		searchH:    searchH,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// Start 启动定时任务
func (c *CronJob) Start() {
	log.Println("Starting cron jobs...")

	// 启动清理过期缓冲区数据的定时任务（每小时执行一次）
	go c.startBufferCleanupJob()

	log.Println("Cron jobs started successfully")
}

// Stop 停止定时任务
func (c *CronJob) Stop() {
	log.Println("Stopping cron jobs...")
	c.cancelFunc()
	log.Println("Cron jobs stopped")
}

// startBufferCleanupJob 启动清理过期缓冲区数据的定时任务
func (c *CronJob) startBufferCleanupJob() {
	// 立即执行一次
	c.cleanupExpiredBuffers()

	// 然后每小时执行一次
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpiredBuffers()
		case <-c.ctx.Done():
			log.Println("Buffer cleanup job stopped")
			return
		}
	}
}

// cleanupExpiredBuffers 清理过期的缓冲区数据
func (c *CronJob) cleanupExpiredBuffers() {
	log.Println("Starting buffer cleanup...")

	rowsAffected, err := c.searchH.CleanupExpiredBuffers()
	if err != nil {
		log.Printf("Failed to cleanup expired buffers: %v\n", err)
		return
	}

	log.Printf("Buffer cleanup completed. Deleted %d expired records.\n", rowsAffected)
}
