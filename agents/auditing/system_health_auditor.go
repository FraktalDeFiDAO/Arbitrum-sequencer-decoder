// Package auditing provides auditing agents for the arbitrum-sequencer-decoder system
package auditing

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// SystemHealthAuditor monitors the overall health of the system
// Following the project's audit framework, security guidelines, and performance recommendations
type SystemHealthAuditor struct {
	mu           sync.RWMutex
	stats        *SystemHealthStats
	notification chan AuditEvent
	config       *AuditConfig
	healthChecks map[string]HealthCheckFunc
	components   map[string]*ComponentStatus

	// Additional metrics and monitoring following audit recommendations
	securityMetrics    *SystemHealthSecurityMetrics
	performanceMetrics *SystemHealthPerformanceMetrics
	auditTrail         []AuditEvent // For audit trail following audit framework
}

// Maximum size for audit trail to prevent unbounded memory growth
const maxAuditTrailSize = 10000

// SystemHealthStats holds statistics about the system's health
type SystemHealthStats struct {
	StartTime           time.Time
	ActiveDecoders      int
	ActivePools         int64
	ActiveOpportunities int64
	TotalTransactions   int64
	AvgLatency          time.Duration
	ErrorRate           float64
	LastHealthCheck     time.Time
}

// SystemHealthPerformanceMetrics holds system health-specific performance metrics
type SystemHealthPerformanceMetrics struct {
	MaxCheckTime   time.Duration
	MinCheckTime   time.Duration
	TotalCheckTime time.Duration
	CheckCount     int64
}

// SystemHealthSecurityMetrics holds system health-specific security metrics
type SystemHealthSecurityMetrics struct {
	RiskEventsDetected   int64
	InvalidOpportunities int64
	SuspiciousActivities int64
	LastSecurityEvent    time.Time
}

// ComponentStatus represents the status of a system component
type ComponentStatus struct {
	Name         string
	Status       string // "healthy", "degraded", "unhealthy", "unknown"
	LastUpdate   time.Time
	ErrorMessage string
	ResponseTime time.Duration
	LastCheck    time.Time
}

// HealthCheckFunc is a function that performs a health check
type HealthCheckFunc func() (bool, error, time.Duration)

// NewSystemHealthAuditor creates a new system health auditor
// Following security best practices and performance optimization guidelines
func NewSystemHealthAuditor(config *AuditConfig) *SystemHealthAuditor {
	if config == nil {
		config = &AuditConfig{
			Interval:       30 * time.Second,
			EnableStats:    true,
			EnableAlerts:   true,
			MinSuccessRate: 0.95,
			MaxDecodeTime:  100 * time.Millisecond,
		}
	}

	auditor := &SystemHealthAuditor{
		stats: &SystemHealthStats{
			StartTime: time.Now(),
		},
		notification: make(chan AuditEvent, 100),
		config:       config,
		healthChecks: make(map[string]HealthCheckFunc),
		components:   make(map[string]*ComponentStatus),
		securityMetrics: &SystemHealthSecurityMetrics{
			RiskEventsDetected:   0,
			InvalidOpportunities: 0,
			SuspiciousActivities: 0,
			LastSecurityEvent:    time.Time{}, // Zero time initially
		},
		performanceMetrics: &SystemHealthPerformanceMetrics{
			MaxCheckTime:   0,
			MinCheckTime:   time.Duration(1<<63 - 1), // Max possible duration
			TotalCheckTime: 0,
			CheckCount:     0,
		},
		auditTrail: make([]AuditEvent, 0),
	}

	// Initialize default components
	auditor.components["sequencer_reader"] = &ComponentStatus{Name: "sequencer_reader", Status: "unknown", LastUpdate: time.Now()}
	auditor.components["decoder_manager"] = &ComponentStatus{Name: "decoder_manager", Status: "unknown", LastUpdate: time.Now()}
	auditor.components["pool_oracle"] = &ComponentStatus{Name: "pool_oracle", Status: "unknown", LastUpdate: time.Now()}
	auditor.components["arb_engine"] = &ComponentStatus{Name: "arb_engine", Status: "unknown", LastUpdate: time.Now()}
	auditor.components["transaction_executor"] = &ComponentStatus{Name: "transaction_executor", Status: "unknown", LastUpdate: time.Now()}

	return auditor
}

// Start begins auditing system health in a separate goroutine
func (sha *SystemHealthAuditor) Start(ctx context.Context) error {
	log.Println("Starting system health auditor...")

	// Start a goroutine to periodically check component health
	go sha.checkComponentHealth(ctx)

	// Start a goroutine to periodically report stats
	go sha.reportStats(ctx)

	// Start a goroutine to handle notifications
	go sha.handleNotifications(ctx)

	// The auditor is now running
	<-ctx.Done()
	log.Println("System health auditor stopped")
	return ctx.Err()
}

// RegisterHealthCheck registers a health check function for a component
func (sha *SystemHealthAuditor) RegisterHealthCheck(componentName string, checkFunc HealthCheckFunc) {
	sha.mu.Lock()
	defer sha.mu.Unlock()

	sha.healthChecks[componentName] = checkFunc
	if _, exists := sha.components[componentName]; !exists {
		sha.components[componentName] = &ComponentStatus{
			Name:       componentName,
			Status:     "unknown",
			LastUpdate: time.Now(),
		}
	}
}

// UpdateStats updates the system health statistics
func (sha *SystemHealthAuditor) UpdateStats(updateFunc func(*SystemHealthStats)) {
	sha.mu.Lock()
	defer sha.mu.Unlock()

	updateFunc(sha.stats)
	sha.stats.LastHealthCheck = time.Now()
}

// GetStats returns the current system health statistics
func (sha *SystemHealthAuditor) GetStats() *SystemHealthStats {
	sha.mu.RLock()
	defer sha.mu.RUnlock()

	return &SystemHealthStats{
		StartTime:           sha.stats.StartTime,
		ActiveDecoders:      sha.stats.ActiveDecoders,
		ActivePools:         sha.stats.ActivePools,
		ActiveOpportunities: sha.stats.ActiveOpportunities,
		TotalTransactions:   sha.stats.TotalTransactions,
		AvgLatency:          sha.stats.AvgLatency,
		ErrorRate:           sha.stats.ErrorRate,
		LastHealthCheck:     sha.stats.LastHealthCheck,
	}
}

// GetComponentStatus returns the status of a specific component
func (sha *SystemHealthAuditor) GetComponentStatus(componentName string) *ComponentStatus {
	sha.mu.RLock()
	defer sha.mu.RUnlock()

	if status, exists := sha.components[componentName]; exists {
		return &ComponentStatus{
			Name:         status.Name,
			Status:       status.Status,
			LastUpdate:   status.LastUpdate,
			ErrorMessage: status.ErrorMessage,
			ResponseTime: status.ResponseTime,
			LastCheck:    status.LastCheck,
		}
	}
	return nil
}

// GetAllComponentStatuses returns the status of all components
func (sha *SystemHealthAuditor) GetAllComponentStatuses() map[string]*ComponentStatus {
	sha.mu.RLock()
	defer sha.mu.RUnlock()

	result := make(map[string]*ComponentStatus)
	for name, status := range sha.components {
		result[name] = &ComponentStatus{
			Name:         status.Name,
			Status:       status.Status,
			LastUpdate:   status.LastUpdate,
			ErrorMessage: status.ErrorMessage,
			ResponseTime: status.ResponseTime,
			LastCheck:    status.LastCheck,
		}
	}
	return result
}

// checkComponentHealth periodically checks the health of registered components
func (sha *SystemHealthAuditor) checkComponentHealth(ctx context.Context) {
	ticker := time.NewTicker(sha.config.Interval / 2) // More frequent checks
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sha.performHealthChecks()
		}
	}
}

// performHealthChecks executes all registered health checks
// Following audit framework and performance monitoring recommendations
func (sha *SystemHealthAuditor) performHealthChecks() {
	sha.mu.Lock()
	defer sha.mu.Unlock()

	startTime := time.Now()

	for componentName, checkFunc := range sha.healthChecks {
		isHealthy, err, responseTime := checkFunc()

		status := &ComponentStatus{
			Name:         componentName,
			Status:       "unknown",
			LastUpdate:   time.Now(),
			ErrorMessage: "",
			ResponseTime: responseTime,
			LastCheck:    time.Now(),
		}

		if err != nil {
			status.Status = "unhealthy"
			status.ErrorMessage = err.Error()
		} else if isHealthy {
			status.Status = "healthy"
		} else {
			status.Status = "degraded"
		}

		// Update the component status
		sha.components[componentName] = status

		// Add to audit trail following audit framework
		auditEvent := AuditEvent{
			Timestamp: time.Now(),
			EventType: "component_health_check",
			Decoder:   componentName,
			Message:   fmt.Sprintf("Component %s health check completed", componentName),
			Data: map[string]interface{}{
				"status":        status.Status,
				"response_time": status.ResponseTime,
				"error":         status.ErrorMessage,
			},
			Error: err,
		}
		sha.auditTrail = append(sha.auditTrail, auditEvent)

		// Trim audit trail if it exceeds max size (keep most recent events)
		if len(sha.auditTrail) > maxAuditTrailSize {
			sha.auditTrail = sha.auditTrail[len(sha.auditTrail)-maxAuditTrailSize:]
		}

		// Send notification if status changed or if it's unhealthy
		if status.Status == "unhealthy" || status.Status == "degraded" {
			event := AuditEvent{
				Timestamp: time.Now(),
				EventType: "component_health_issue",
				Decoder:   componentName,
				Message:   fmt.Sprintf("Component %s is %s: %s", componentName, status.Status, status.ErrorMessage),
				Data: map[string]interface{}{
					"status":        status.Status,
					"response_time": status.ResponseTime,
					"error":         status.ErrorMessage,
				},
			}

			// Send notification non-blockingly
			select {
			case sha.notification <- event:
			default:
				log.Printf("Audit notification channel full, dropping health event for component %s", componentName)
			}

			// Update security metrics for failed checks
			sha.securityMetrics.SuspiciousActivities++
			sha.securityMetrics.LastSecurityEvent = time.Now()
		}
	}

	// Update performance metrics
	checkDuration := time.Since(startTime)
	if checkDuration > sha.performanceMetrics.MaxCheckTime {
		sha.performanceMetrics.MaxCheckTime = checkDuration
	}
	if checkDuration < sha.performanceMetrics.MinCheckTime && checkDuration > 0 {
		sha.performanceMetrics.MinCheckTime = checkDuration
	}
	sha.performanceMetrics.TotalCheckTime += checkDuration
	sha.performanceMetrics.CheckCount++
}

// reportStats periodically reports system health statistics
func (sha *SystemHealthAuditor) reportStats(ctx context.Context) {
	ticker := time.NewTicker(sha.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := sha.GetStats()
			uptime := time.Since(stats.StartTime)

			log.Printf("System Health Stats: Uptime=%v, ActiveDecoders=%d, ActivePools=%d, ActiveOpportunities=%d, TotalTransactions=%d, AvgLatency=%v, ErrorRate=%.2f%%",
				uptime, stats.ActiveDecoders, stats.ActivePools, stats.ActiveOpportunities, stats.TotalTransactions, stats.AvgLatency, stats.ErrorRate*100)

			// Report component statuses
			components := sha.GetAllComponentStatuses()
			for name, status := range components {
				log.Printf("Component %s: %s (ResponseTime: %v, LastCheck: %v)",
					name, status.Status, status.ResponseTime, status.LastCheck)
			}
		}
	}
}

// handleNotifications processes system health audit notifications
func (sha *SystemHealthAuditor) handleNotifications(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-sha.notification:
			// Log the event
			log.Printf("SYSTEM HEALTH AUDIT EVENT: %s - Component: %s, Message: %s",
				event.EventType, event.Decoder, event.Message)

			// Handle specific event types
			if sha.config.EnableAlerts {
				switch event.EventType {
				case "component_health_issue":
					alertEvent := AuditEvent{
						Timestamp: time.Now(),
						EventType: "alert_component_issue",
						Decoder:   "system_health",
						Message:   fmt.Sprintf("Critical component issue detected: %s", event.Message),
						Data:      event.Data,
					}
					select {
					case sha.notification <- alertEvent:
					default:
						log.Println("Alert notification channel full")
					}
				}
			}
		}
	}
}

// MonitorPoolOracle adds health monitoring for a PoolOracle component
func (sha *SystemHealthAuditor) MonitorPoolOracle(oracle types.PoolOracle) {
	sha.RegisterHealthCheck("pool_oracle", func() (bool, error, time.Duration) {
		start := time.Now()

		// Perform a simple health check by getting a known pool
		// For this example, we'll just check if oracle is not nil
		// In a real implementation, you might check if you can access the actual pool data
		healthy := oracle != nil

		if !healthy {
			return false, fmt.Errorf("pool oracle is not initialized"), time.Since(start)
		}

		// Update stats with active pool count if possible
		// This is a simplified check - in reality, you'd have a method to get the count
		sha.UpdateStats(func(stats *SystemHealthStats) {
			// stats.ActivePools = count of pools
		})

		return true, nil, time.Since(start)
	})
}

// MonitorDecoder adds health monitoring for a Decoder component
func (sha *SystemHealthAuditor) MonitorDecoder(decoder types.Decoder) {
	name := string(decoder.Protocol())
	sha.RegisterHealthCheck(name, func() (bool, error, time.Duration) {
		start := time.Now()

		// Perform a simple health check by verifying the decoder is not nil
		healthy := decoder != nil

		if !healthy {
			return false, fmt.Errorf("decoder is not initialized"), time.Since(start)
		}

		return true, nil, time.Since(start)
	})
}

// MonitorArbitrageEngine adds health monitoring for an ArbitrageEngine component
func (sha *SystemHealthAuditor) MonitorArbitrageEngine(engine types.ArbitrageEngine) {
	sha.RegisterHealthCheck("arb_engine", func() (bool, error, time.Duration) {
		start := time.Now()

		// Perform a simple health check by verifying the engine is not nil
		healthy := engine != nil

		if !healthy {
			return false, fmt.Errorf("arbitrage engine is not initialized"), time.Since(start)
		}

		return true, nil, time.Since(start)
	})
}

// MonitorTransactionExecutor adds health monitoring for a TransactionExecutor component
func (sha *SystemHealthAuditor) MonitorTransactionExecutor(executor types.TransactionExecutor) {
	sha.RegisterHealthCheck("transaction_executor", func() (bool, error, time.Duration) {
		start := time.Now()

		// Perform a simple health check by verifying the executor is not nil
		healthy := executor != nil

		if !healthy {
			return false, fmt.Errorf("transaction executor is not initialized"), time.Since(start)
		}

		return true, nil, time.Since(start)
	})
}

// GetSecurityMetrics returns security metrics for monitoring
// Following security audit recommendations
func (sha *SystemHealthAuditor) GetSecurityMetrics() *SystemHealthSecurityMetrics {
	sha.mu.RLock()
	defer sha.mu.RUnlock()

	return &SystemHealthSecurityMetrics{
		RiskEventsDetected:   sha.securityMetrics.RiskEventsDetected,
		InvalidOpportunities: sha.securityMetrics.InvalidOpportunities,
		SuspiciousActivities: sha.securityMetrics.SuspiciousActivities,
		LastSecurityEvent:    sha.securityMetrics.LastSecurityEvent,
	}
}

// GetPerformanceMetrics returns performance metrics for monitoring
// Following performance audit recommendations
func (sha *SystemHealthAuditor) GetPerformanceMetrics() *SystemHealthPerformanceMetrics {
	sha.mu.RLock()
	defer sha.mu.RUnlock()

	if sha.performanceMetrics.CheckCount == 0 {
		return &SystemHealthPerformanceMetrics{
			MaxCheckTime:   0,
			MinCheckTime:   0,
			TotalCheckTime: 0,
			CheckCount:     0,
		}
	}

	return &SystemHealthPerformanceMetrics{
		MaxCheckTime:   sha.performanceMetrics.MaxCheckTime,
		MinCheckTime:   sha.performanceMetrics.MinCheckTime,
		TotalCheckTime: sha.performanceMetrics.TotalCheckTime,
		CheckCount:     sha.performanceMetrics.CheckCount,
	}
}

// GetAuditTrail returns the audit trail for review
// Following audit framework recommendations
func (sha *SystemHealthAuditor) GetAuditTrail() []AuditEvent {
	sha.mu.RLock()
	defer sha.mu.RUnlock()

	// Return a copy to prevent external modifications
	trail := make([]AuditEvent, len(sha.auditTrail))
	copy(trail, sha.auditTrail)

	return trail
}

// GetNotificationChannel returns the channel for audit notifications
func (sha *SystemHealthAuditor) GetNotificationChannel() <-chan AuditEvent {
	return sha.notification
}
