// Package auditing provides auditing agents for the arbitrum-sequencer-decoder system
package auditing

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/FraktalDeFiDAO/Arbitrum-sequencer-decoder/pkg/types"
)

// ArbitragePerformanceMetrics holds arbitrage auditor-specific performance metrics
type ArbitragePerformanceMetrics struct {
	MaxValidationTime   time.Duration
	MinValidationTime   time.Duration
	TotalValidationTime time.Duration
	ValidationCount     int64
}

// ArbitrageAuditor audits arbitrage opportunities to ensure they are valid and profitable
// Following the project's audit framework, security guidelines, and performance recommendations
type ArbitrageAuditor struct {
	mu              sync.RWMutex
	stats           *ArbitrageStats
	arbEngine       types.ArbitrageEngine
	notification    chan AuditEvent
	config          *AuditConfig
	validationCache map[string]*ValidationResult // Cache validation results

	// Additional security and monitoring fields
	securityMetrics    *ArbitrageSecurityMetrics
	performanceMetrics *ArbitragePerformanceMetrics
}

// ArbitrageSecurityMetrics holds arbitrage auditor-specific security metrics
type ArbitrageSecurityMetrics struct {
	RiskEventsDetected   int64
	InvalidOpportunities int64
	SuspiciousActivities int64
	LastSecurityEvent    time.Time
}

// ArbitrageStats holds statistics about arbitrage opportunities
type ArbitrageStats struct {
	TotalOpportunities int64
	Validated          int64
	Invalid            int64
	Executed           int64
	Unprofitable       int64
	Expired            int64
	AvgProfit          *big.Float
	AvgRisk            float64
	LastUpdated        time.Time
}

// ValidationResult holds the result of validating an arbitrage opportunity
type ValidationResult struct {
	IsValid        bool
	Profitability  *big.Float
	RiskScore      float64
	Reason         string
	ValidationTime time.Time
	Expiration     time.Time
}

// NewArbitrageAuditor creates a new arbitrage auditor
// Following security best practices and performance optimization guidelines
func NewArbitrageAuditor(arbEngine types.ArbitrageEngine, config *AuditConfig) *ArbitrageAuditor {
	if config == nil {
		config = &AuditConfig{
			Interval:       30 * time.Second,
			EnableStats:    true,
			EnableAlerts:   true,
			MinSuccessRate: 0.95,
			MaxDecodeTime:  100 * time.Millisecond,
		}
	}

	auditor := &ArbitrageAuditor{
		stats: &ArbitrageStats{
			TotalOpportunities: 0,
			Validated:          0,
			Invalid:            0,
			Executed:           0,
			Unprofitable:       0,
			Expired:            0,
			AvgProfit:          big.NewFloat(0),
			AvgRisk:            0.0,
			LastUpdated:        time.Now(),
		},
		arbEngine:       arbEngine,
		notification:    make(chan AuditEvent, 100),
		config:          config,
		validationCache: make(map[string]*ValidationResult),
		securityMetrics: &ArbitrageSecurityMetrics{
			RiskEventsDetected:   0,
			InvalidOpportunities: 0,
			SuspiciousActivities: 0,
			LastSecurityEvent:    time.Time{}, // Zero time initially
		},
		performanceMetrics: &ArbitragePerformanceMetrics{
			MaxValidationTime:   0,
			MinValidationTime:   time.Duration(1<<63 - 1), // Max possible duration
			TotalValidationTime: 0,
			ValidationCount:     0,
		},
	}

	return auditor
}

// Start begins auditing arbitrage opportunities in a separate goroutine
func (aa *ArbitrageAuditor) Start(ctx context.Context) error {
	log.Println("Starting arbitrage auditor...")

	// Start a goroutine to periodically report stats
	go aa.reportStats(ctx)

	// Start a goroutine to handle notifications
	go aa.handleNotifications(ctx)

	// Start a goroutine to clean the validation cache
	go aa.cleanCache(ctx)

	// The auditor is now running
	<-ctx.Done()
	log.Println("Arbitrage auditor stopped")
	return ctx.Err()
}

// AuditOpportunity audits a single arbitrage opportunity
// Following security best practices and performance monitoring recommendations
func (aa *ArbitrageAuditor) AuditOpportunity(opportunity types.ArbitrageOpportunity) (*ValidationResult, error) {
	startTime := time.Now()

	validationResult, err := aa.validateOpportunity(opportunity)
	if err != nil {
		return nil, fmt.Errorf("failed to validate opportunity: %w", err)
	}

	validationTime := time.Since(startTime)

	// Update performance metrics
	aa.updatePerformanceMetrics(validationTime)

	// Update security metrics if needed
	if !validationResult.IsValid {
		aa.updateSecurityMetrics()
	}

	// Update stats
	aa.updateStats(&opportunity, validationResult)

	// Send notification
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: "arbitrage_validation",
		Decoder:   string(opportunity.Type),
		Data: map[string]interface{}{
			"opportunity_id":     opportunity.ID,
			"validation_time":    validationTime,
			"is_valid":           validationResult.IsValid,
			"profitability":      validationResult.Profitability,
			"risk_score":         validationResult.RiskScore,
			"execution_path_len": len(opportunity.ExecutionPath),
			"type":               opportunity.Type,
		},
		Message: validationResult.Reason,
	}

	// Send notification non-blockingly
	select {
	case aa.notification <- event:
	default:
		log.Printf("Audit notification channel full, dropping event for opportunity %s", opportunity.ID)
	}

	return validationResult, nil
}

// updatePerformanceMetrics updates the performance metrics tracking
func (aa *ArbitrageAuditor) updatePerformanceMetrics(validationTime time.Duration) {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	// Update validation time metrics
	if validationTime > aa.performanceMetrics.MaxValidationTime {
		aa.performanceMetrics.MaxValidationTime = validationTime
	}
	if validationTime < aa.performanceMetrics.MinValidationTime && validationTime > 0 {
		aa.performanceMetrics.MinValidationTime = validationTime
	}

	aa.performanceMetrics.TotalValidationTime += validationTime
	aa.performanceMetrics.ValidationCount++
}

// updateSecurityMetrics updates the security metrics tracking
func (aa *ArbitrageAuditor) updateSecurityMetrics() {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	aa.securityMetrics.InvalidOpportunities++
	aa.securityMetrics.LastSecurityEvent = time.Now()
}

// GetSecurityMetrics returns security metrics for monitoring
// Following security audit recommendations
func (aa *ArbitrageAuditor) GetSecurityMetrics() *ArbitrageSecurityMetrics {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	return &ArbitrageSecurityMetrics{
		RiskEventsDetected:   aa.securityMetrics.RiskEventsDetected,
		InvalidOpportunities: aa.securityMetrics.InvalidOpportunities,
		SuspiciousActivities: aa.securityMetrics.SuspiciousActivities,
		LastSecurityEvent:    aa.securityMetrics.LastSecurityEvent,
	}
}

// GetPerformanceMetrics returns performance metrics for monitoring
// Following performance audit recommendations
func (aa *ArbitrageAuditor) GetPerformanceMetrics() *ArbitragePerformanceMetrics {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	if aa.performanceMetrics.ValidationCount == 0 {
		return &ArbitragePerformanceMetrics{
			MaxValidationTime:   0,
			MinValidationTime:   0,
			TotalValidationTime: 0,
			ValidationCount:     0,
		}
	}

	return &ArbitragePerformanceMetrics{
		MaxValidationTime:   aa.performanceMetrics.MaxValidationTime,
		MinValidationTime:   aa.performanceMetrics.MinValidationTime,
		TotalValidationTime: aa.performanceMetrics.TotalValidationTime,
		ValidationCount:     aa.performanceMetrics.ValidationCount,
	}
}

// validateOpportunity performs validation of an arbitrage opportunity
func (aa *ArbitrageAuditor) validateOpportunity(opportunity types.ArbitrageOpportunity) (*ValidationResult, error) {
	// Check if opportunity is expired
	if time.Now().After(opportunity.Expiration) {
		return &ValidationResult{
			IsValid:        false,
			Profitability:  big.NewFloat(0),
			RiskScore:      1.0,
			Reason:         "Opportunity has expired",
			ValidationTime: time.Now(),
			Expiration:     opportunity.Expiration,
		}, nil
	}

	// Calculate profitability (considering opportunity.Profit)
	profitability := new(big.Float).SetInt(opportunity.Profit)

	// Calculate risk score based on probability and risk factor
	riskScore := (1.0 - opportunity.Probability) * 0.5 // Lower probability = higher risk
	riskScore += opportunity.RiskFactor * 0.5          // Add inherent risk factor

	isValid := riskScore < 0.7 && profitability.Sign() > 0 // Consider valid if risk < 70% and profitable

	reason := "Valid opportunity"
	if riskScore >= 0.7 {
		reason = "High risk opportunity"
	} else if profitability.Sign() <= 0 {
		reason = "Not profitable"
	}

	result := &ValidationResult{
		IsValid:        isValid,
		Profitability:  profitability,
		RiskScore:      riskScore,
		Reason:         reason,
		ValidationTime: time.Now(),
		Expiration:     opportunity.Expiration,
	}

	// Cache this validation result
	aa.mu.Lock()
	aa.validationCache[opportunity.ID] = result
	aa.mu.Unlock()

	return result, nil
}

// updateStats updates the statistics for arbitrage opportunities
func (aa *ArbitrageAuditor) updateStats(opportunity *types.ArbitrageOpportunity, result *ValidationResult) {
	aa.mu.Lock()
	defer aa.mu.Unlock()

	aa.stats.TotalOpportunities++
	if result.IsValid {
		aa.stats.Validated++
	} else {
		aa.stats.Invalid++
	}

	// Update average profitability
	totalOpportunities := aa.stats.Validated + aa.stats.Invalid
	avgProfitFloat, _ := aa.stats.AvgProfit.Float64()
	newAvgProfit := new(big.Float).Quo(
		new(big.Float).Add(
			new(big.Float).Mul(big.NewFloat(avgProfitFloat), big.NewFloat(float64(totalOpportunities-1))),
			result.Profitability,
		),
		big.NewFloat(float64(totalOpportunities)),
	)
	aa.stats.AvgProfit = newAvgProfit

	// Update average risk
	avgRisk := aa.stats.AvgRisk
	newAvgRisk := (avgRisk*float64(totalOpportunities-1) + result.RiskScore) / float64(totalOpportunities)
	aa.stats.AvgRisk = newAvgRisk

	aa.stats.LastUpdated = time.Now()
}

// GetStats returns the current arbitrage statistics
func (aa *ArbitrageAuditor) GetStats() *ArbitrageStats {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	return &ArbitrageStats{
		TotalOpportunities: aa.stats.TotalOpportunities,
		Validated:          aa.stats.Validated,
		Invalid:            aa.stats.Invalid,
		Executed:           aa.stats.Executed,
		Unprofitable:       aa.stats.Unprofitable,
		AvgProfit:          new(big.Float).Set(aa.stats.AvgProfit),
		AvgRisk:            aa.stats.AvgRisk,
		LastUpdated:        aa.stats.LastUpdated,
	}
}

// reportStats periodically reports audit statistics
func (aa *ArbitrageAuditor) reportStats(ctx context.Context) {
	ticker := time.NewTicker(aa.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := aa.GetStats()
			avgProfitNormalized, _ := new(big.Float).Quo(stats.AvgProfit, big.NewFloat(1e18)).Float64()
			log.Printf("Arbitrage Auditor Stats: Total=%d, Validated=%d, Invalid=%d, AvgProfit=%.2f, AvgRisk=%.2f%%",
				stats.TotalOpportunities, stats.Validated, stats.Invalid,
				avgProfitNormalized,
				stats.AvgRisk*100)
		}
	}
}

// handleNotifications processes arbitrage audit notifications
func (aa *ArbitrageAuditor) handleNotifications(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-aa.notification:
			// Log the event - safely extract is_valid from Data
			isValid := false
			if eventData, ok := event.Data.(map[string]interface{}); ok {
				if val, exists := eventData["is_valid"]; exists {
					if b, ok := val.(bool); ok {
						isValid = b
					}
				}
			}
			log.Printf("ARBITRAGE AUDIT EVENT: %s - Type: %s, Valid: %v, Message: %s",
				event.EventType, event.Decoder, isValid, event.Message)

			// Handle specific event types
			if aa.config.EnableAlerts {
				switch event.EventType {
				case "high_risk_opportunity":
					alertEvent := AuditEvent{
						Timestamp: time.Now(),
						EventType: "alert_high_risk",
						Decoder:   "arbitrage",
						Message:   fmt.Sprintf("High risk opportunity detected: %s", event.Message),
						Data:      event.Data,
					}
					select {
					case aa.notification <- alertEvent:
					default:
						log.Println("Alert notification channel full")
					}
				}
			}
		}
	}
}

// cleanCache periodically cleans expired validation results from the cache
func (aa *ArbitrageAuditor) cleanCache(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Clean every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			aa.mu.Lock()
			for id, result := range aa.validationCache {
				if time.Now().After(result.Expiration) {
					delete(aa.validationCache, id)
				}
			}
			aa.mu.Unlock()
		}
	}
}

// ValidateAndExecute validates an arbitrage opportunity and executes if valid
func (aa *ArbitrageAuditor) ValidateAndExecute(executor types.TransactionExecutor, opportunity types.ArbitrageOpportunity) error {
	validationResult, err := aa.AuditOpportunity(opportunity)
	if err != nil {
		return fmt.Errorf("failed to audit opportunity: %w", err)
	}

	if !validationResult.IsValid {
		return fmt.Errorf("opportunity is invalid: %s", validationResult.Reason)
	}

	// Verify opportunity is still valid just before execution
	isValid, err := executor.Validate(opportunity)
	if err != nil {
		return fmt.Errorf("failed to re-validate opportunity before execution: %w", err)
	}

	if !isValid {
		return fmt.Errorf("opportunity became invalid before execution")
	}

	// Execute the opportunity
	if err := executor.Execute(opportunity); err != nil {
		return fmt.Errorf("failed to execute opportunity: %w", err)
	}

	// Update stats to reflect execution
	aa.mu.Lock()
	aa.stats.Executed++
	aa.mu.Unlock()

	log.Printf("Successfully executed arbitrage opportunity: %s", opportunity.ID)

	return nil
}

// GetNotificationChannel returns the channel for audit notifications
func (aa *ArbitrageAuditor) GetNotificationChannel() <-chan AuditEvent {
	return aa.notification
}
