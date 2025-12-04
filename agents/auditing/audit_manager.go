// Package auditing provides auditing agents for the arbitrum-sequencer-decoder system
package auditing

import (
	"context"
	"sync"
	"time"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// AuditManager coordinates all auditing activities in accordance with the project's audit framework
type AuditManager struct {
	mu                  sync.RWMutex
	decoderAuditor      *DecoderAuditor
	arbitrageAuditor    *ArbitrageAuditor
	systemHealthAuditor *SystemHealthAuditor
	config              *AuditConfig

	// For code quality: ensure thread-safe operations
	statsMu      sync.RWMutex
	auditMetrics *AuditMetrics
}

// AuditMetrics holds aggregated metrics from all auditors
type AuditMetrics struct {
	StartTime            int64 // Unix timestamp for uptime calculation
	TotalAuditsProcessed int64
	AlertsGenerated      int64
	ErrorsDetected       int64
	PerformanceMetrics   map[string]interface{} // For performance monitoring
}

// NewAuditManager creates a new audit manager that coordinates all auditing agents
func NewAuditManager(config *AuditConfig) *AuditManager {
	if config == nil {
		config = &AuditConfig{
			Interval:       30 * time.Second,
			EnableStats:    true,
			EnableAlerts:   true,
			MinSuccessRate: 0.95,
			MaxDecodeTime:  100 * time.Millisecond,
		}
	}

	auditManager := &AuditManager{
		config: config,
		auditMetrics: &AuditMetrics{
			StartTime:            0, // Will be set when started
			TotalAuditsProcessed: 0,
			AlertsGenerated:      0,
			ErrorsDetected:       0,
			PerformanceMetrics:   make(map[string]interface{}),
		},
	}

	return auditManager
}

// InitializeAuditors sets up all auditing components with their dependencies
// Following security best practices and performance optimization recommendations
func (am *AuditManager) InitializeAuditors(decoder types.Decoder, oracle types.PoolOracle, arbEngine types.ArbitrageEngine, executor types.TransactionExecutor) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.decoderAuditor = NewDecoderAuditor(decoder, oracle, am.config)
	am.arbitrageAuditor = NewArbitrageAuditor(arbEngine, am.config)
	am.systemHealthAuditor = NewSystemHealthAuditor(am.config)

	// Register health checks for components following security recommendations
	am.systemHealthAuditor.MonitorDecoder(decoder)
	am.systemHealthAuditor.MonitorPoolOracle(oracle)
	am.systemHealthAuditor.MonitorArbitrageEngine(arbEngine)
	am.systemHealthAuditor.MonitorTransactionExecutor(executor)

	// Initialize start time for uptime tracking
	am.statsMu.Lock()
	am.auditMetrics.StartTime = time.Now().Unix()
	am.statsMu.Unlock()
}

// Start begins all auditing processes with comprehensive error handling
func (am *AuditManager) Start(ctx context.Context) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Initialize metrics start time if not already set
	am.statsMu.Lock()
	if am.auditMetrics.StartTime == 0 {
		am.auditMetrics.StartTime = time.Now().Unix()
	}
	am.statsMu.Unlock()

	// Create a context that cancels if any auditor fails
	errGroup := make(chan error, 3)

	// Start decoder auditor
	go func() {
		if am.decoderAuditor != nil {
			err := am.decoderAuditor.Start(ctx)
			if err != nil {
				am.updateErrorMetric()
			}
			errGroup <- err
		} else {
			errGroup <- nil
		}
	}()

	// Start arbitrage auditor
	go func() {
		if am.arbitrageAuditor != nil {
			err := am.arbitrageAuditor.Start(ctx)
			if err != nil {
				am.updateErrorMetric()
			}
			errGroup <- err
		} else {
			errGroup <- nil
		}
	}()

	// Start system health auditor
	go func() {
		if am.systemHealthAuditor != nil {
			err := am.systemHealthAuditor.Start(ctx)
			if err != nil {
				am.updateErrorMetric()
			}
			errGroup <- err
		} else {
			errGroup <- nil
		}
	}()

	// Wait for any auditor to fail
	for i := 0; i < 3; i++ {
		if err := <-errGroup; err != nil {
			return err
		}
	}

	return nil
}

// AuditTransaction audits a transaction through the decoder auditor
// Implements the new audit framework requirements
func (am *AuditManager) AuditTransaction(txHash string, decodeFunc func() ([]types.DecodedAction, error)) ([]types.DecodedAction, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Update audit metrics
	am.updateAuditMetric()

	if am.decoderAuditor == nil {
		// If no auditor, just execute the function directly
		return decodeFunc()
	}

	actions, err := am.decoderAuditor.AuditTransaction(txHash, decodeFunc)
	if err != nil {
		am.updateErrorMetric()
	}
	return actions, err
}

// AuditOpportunity audits an arbitrage opportunity
// Following security and validation best practices
func (am *AuditManager) AuditOpportunity(opportunity types.ArbitrageOpportunity) (*ValidationResult, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.arbitrageAuditor == nil {
		return nil, nil
	}

	result, err := am.arbitrageAuditor.AuditOpportunity(opportunity)
	if err != nil {
		am.updateErrorMetric()
	}
	return result, err
}

// ValidateAndExecute validates an arbitrage opportunity and executes it if valid
// Implements security-first approach with multiple validation layers
func (am *AuditManager) ValidateAndExecute(executor types.TransactionExecutor, opportunity types.ArbitrageOpportunity) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.arbitrageAuditor == nil {
		return executor.Execute(opportunity)
	}

	err := am.arbitrageAuditor.ValidateAndExecute(executor, opportunity)
	if err != nil {
		am.updateErrorMetric()
	}
	return err
}

// UpdateSystemStats updates the system health statistics
func (am *AuditManager) UpdateSystemStats(updateFunc func(*SystemHealthStats)) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.systemHealthAuditor != nil {
		am.systemHealthAuditor.UpdateStats(updateFunc)
	}
}

// GetDecoderStats returns decoder statistics
func (am *AuditManager) GetDecoderStats() map[string]*DecoderStats {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.decoderAuditor != nil {
		return am.decoderAuditor.GetStats()
	}
	return nil
}

// GetArbitrageStats returns arbitrage statistics
func (am *AuditManager) GetArbitrageStats() *ArbitrageStats {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.arbitrageAuditor != nil {
		return am.arbitrageAuditor.GetStats()
	}
	return nil
}

// GetSystemHealthStats returns system health statistics
func (am *AuditManager) GetSystemHealthStats() *SystemHealthStats {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.systemHealthAuditor != nil {
		return am.systemHealthAuditor.GetStats()
	}
	return nil
}

// GetComponentStatus returns the status of a system component
func (am *AuditManager) GetComponentStatus(componentName string) *ComponentStatus {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.systemHealthAuditor != nil {
		return am.systemHealthAuditor.GetComponentStatus(componentName)
	}
	return nil
}

// GetAllComponentStatuses returns the status of all components
func (am *AuditManager) GetAllComponentStatuses() map[string]*ComponentStatus {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.systemHealthAuditor != nil {
		return am.systemHealthAuditor.GetAllComponentStatuses()
	}
	return nil
}

// GetAuditMetrics returns aggregated audit metrics
// Implements the observability infrastructure recommendations
func (am *AuditManager) GetAuditMetrics() *AuditMetrics {
	am.statsMu.RLock()
	defer am.statsMu.RUnlock()

	metrics := &AuditMetrics{
		StartTime:            am.auditMetrics.StartTime,
		TotalAuditsProcessed: am.auditMetrics.TotalAuditsProcessed,
		AlertsGenerated:      am.auditMetrics.AlertsGenerated,
		ErrorsDetected:       am.auditMetrics.ErrorsDetected,
		PerformanceMetrics:   make(map[string]interface{}),
	}

	// Copy performance metrics
	for k, v := range am.auditMetrics.PerformanceMetrics {
		metrics.PerformanceMetrics[k] = v
	}

	return metrics
}

// updateAuditMetric safely updates the audit metrics counter
func (am *AuditManager) updateAuditMetric() {
	am.statsMu.Lock()
	am.auditMetrics.TotalAuditsProcessed++
	am.statsMu.Unlock()
}

// updateErrorMetric safely updates the error metrics counter
func (am *AuditManager) updateErrorMetric() {
	am.statsMu.Lock()
	am.auditMetrics.ErrorsDetected++
	am.statsMu.Unlock()
}

// updateAlertMetric safely updates the alert metrics counter
func (am *AuditManager) updateAlertMetric() {
	am.statsMu.Lock()
	am.auditMetrics.AlertsGenerated++
	am.statsMu.Unlock()
}
