// Package auditing provides auditing agents for the arbitrum-sequencer-decoder system
package auditing

import (
	"time"
)

// PerformanceMetrics holds performance-related metrics (shared across auditors)
type PerformanceMetrics struct {
	MaxTime   time.Duration
	MinTime   time.Duration
	TotalTime time.Duration
	Count     int64
}

// SecurityMetrics holds security-related metrics (shared across auditors)
type SecurityMetrics struct {
	RiskEventsDetected   int64
	InvalidOpportunities int64
	SuspiciousActivities int64
	LastSecurityEvent    time.Time
}

// NewPerformanceMetrics creates a new PerformanceMetrics with default values
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		MaxTime:   0,
		MinTime:   time.Duration(1<<63 - 1), // Max possible duration
		TotalTime: 0,
		Count:     0,
	}
}

// NewSecurityMetrics creates a new SecurityMetrics with default values
func NewSecurityMetrics() *SecurityMetrics {
	return &SecurityMetrics{
		RiskEventsDetected:   0,
		InvalidOpportunities: 0,
		SuspiciousActivities: 0,
		LastSecurityEvent:    time.Time{}, // Zero time initially
	}
}

// Update updates the performance metrics with a new measurement
func (pm *PerformanceMetrics) Update(duration time.Duration) {
	if duration > pm.MaxTime {
		pm.MaxTime = duration
	}
	if duration < pm.MinTime && duration > 0 {
		pm.MinTime = duration
	}
	pm.TotalTime += duration
	pm.Count++
}

// AverageTime returns the average time or 0 if no measurements
func (pm *PerformanceMetrics) AverageTime() time.Duration {
	if pm.Count == 0 {
		return 0
	}
	return time.Duration(int64(pm.TotalTime) / pm.Count)
}

// Copy creates a deep copy of PerformanceMetrics
func (pm *PerformanceMetrics) Copy() *PerformanceMetrics {
	if pm == nil {
		return nil
	}
	return &PerformanceMetrics{
		MaxTime:   pm.MaxTime,
		MinTime:   pm.MinTime,
		TotalTime: pm.TotalTime,
		Count:     pm.Count,
	}
}

// Copy creates a deep copy of SecurityMetrics
func (sm *SecurityMetrics) Copy() *SecurityMetrics {
	if sm == nil {
		return nil
	}
	return &SecurityMetrics{
		RiskEventsDetected:   sm.RiskEventsDetected,
		InvalidOpportunities: sm.InvalidOpportunities,
		SuspiciousActivities: sm.SuspiciousActivities,
		LastSecurityEvent:    sm.LastSecurityEvent,
	}
}
