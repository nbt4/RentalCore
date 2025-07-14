package repository

import (
	"fmt"
	"go-barcode-webapp/internal/models"
)

// FreeDevicesFromCompletedJobs removes device assignments from jobs with "paid" status
// and sets the device status back to "free"
func (r *JobRepository) FreeDevicesFromCompletedJobs() error {
	// Find all jobs with "paid" status
	var paidJobs []models.Job
	err := r.db.Joins("JOIN status ON jobs.statusID = status.statusID").
		Where("status.status = ?", "paid").
		Find(&paidJobs).Error
	if err != nil {
		return fmt.Errorf("failed to find paid jobs: %v", err)
	}

	if len(paidJobs) == 0 {
		return nil // No paid jobs found
	}

	// Extract job IDs
	var jobIDs []uint
	for _, job := range paidJobs {
		jobIDs = append(jobIDs, job.JobID)
	}

	// Find all devices assigned to these jobs
	var jobDevices []models.JobDevice
	err = r.db.Where("jobID IN ?", jobIDs).Find(&jobDevices).Error
	if err != nil {
		return fmt.Errorf("failed to find devices in paid jobs: %v", err)
	}

	// Start transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove job device assignments
	if len(jobDevices) > 0 {
		err = tx.Where("jobID IN ?", jobIDs).Delete(&models.JobDevice{}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove job device assignments: %v", err)
		}

		// Set device status back to "free"
		var deviceIDs []string
		for _, jd := range jobDevices {
			deviceIDs = append(deviceIDs, jd.DeviceID)
		}

		err = tx.Model(&models.Device{}).
			Where("deviceID IN ?", deviceIDs).
			Update("status", "free").Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update device status: %v", err)
		}
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}