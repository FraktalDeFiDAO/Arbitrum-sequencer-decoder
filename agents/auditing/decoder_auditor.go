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

// DecoderAuditor audits the performance and accuracy of the decoder implementations
// Following the project's audit framework, code quality standards, and security guidelines
type DecoderAuditor struct {
	mu           sync.RWMutex
	stats        map[string]*DecoderStats
	decoder      types.Decoder
	oracle       types.PoolOracle
	notification chan AuditEvent
	config       *AuditConfig

	// Additional metrics tracking following performance recommendations
	performanceMetrics *DecoderPerformanceMetrics
}

// DecoderPerformanceMetrics holds decoder-specific performance metrics
type DecoderPerformanceMetrics struct {
	MaxDecodeTime   time.Duration
	MinDecodeTime   time.Duration
	TotalDecodeTime time.Duration
	DecodeCount     int64
}

// DecoderStats holds statistics about a decoder's performance
// Updated to follow code quality and audit framework recommendations
type DecoderStats struct {
	TotalDecoded   int64
	TotalFailed    int64
	AvgDecodeTime  time.Duration
	LastDecodeTime time.Duration
	SuccessRate    float64
	LastUpdated    time.Time
	TotalLiquidity *big.Int      // Track estimated liquidity processed (using big.Int to prevent overflow)
	MaxDecodeTime  time.Duration // Following performance metrics tracking
	MinDecodeTime  time.Duration // Following performance metrics tracking
	ErrorDetails   []string      // For enhanced error tracking
}

// AuditEvent represents an auditing event
type AuditEvent struct {
	Timestamp time.Time
	EventType string
	Decoder   string
	Message   string
	Data      interface{}
	Error     error
}

// AuditConfig holds configuration for the auditing system
type AuditConfig struct {
	Interval       time.Duration
	EnableStats    bool
	EnableAlerts   bool
	MinSuccessRate float64
	MaxDecodeTime  time.Duration
}

// NewDecoderAuditor creates a new decoder auditor
// Following security best practices and code quality standards
func NewDecoderAuditor(decoder types.Decoder, oracle types.PoolOracle, config *AuditConfig) *DecoderAuditor {
	if config == nil {
		config = &AuditConfig{
			Interval:       30 * time.Second,
			EnableStats:    true,
			EnableAlerts:   true,
			MinSuccessRate: 0.95, // 95%
			MaxDecodeTime:  100 * time.Millisecond,
		}
	}

	auditor := &DecoderAuditor{
		stats:        make(map[string]*DecoderStats),
		decoder:      decoder,
		oracle:       oracle,
		notification: make(chan AuditEvent, 100), // Buffered to prevent blocking
		config:       config,
		performanceMetrics: &DecoderPerformanceMetrics{
			MaxDecodeTime:   0,
			MinDecodeTime:   time.Duration(1<<63 - 1), // Max possible duration
			TotalDecodeTime: 0,
			DecodeCount:     0,
		},
	}

	return auditor
}

// Start begins auditing in a separate goroutine
func (da *DecoderAuditor) Start(ctx context.Context) error {
	log.Println("Starting decoder auditor...")

	// Start a goroutine to periodically report stats
	go da.reportStats(ctx)

	// Start a goroutine to handle notifications
	go da.handleNotifications(ctx)

	// The auditor is now running
	<-ctx.Done()
	log.Println("Decoder auditor stopped")
	return ctx.Err()
}

// AuditTransaction audits a single transaction decoding
func (da *DecoderAuditor) AuditTransaction(txHash string, decodeFunc func() ([]types.DecodedAction, error)) ([]types.DecodedAction, error) {
	startTime := time.Now()

	actions, err := decodeFunc()

	decodeTime := time.Since(startTime)

	// Update stats
	decoderName := string(da.decoder.Protocol())
	da.updateStats(decoderName, err == nil, decodeTime, actions)

	// Send notification if needed
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: "decode_result",
		Decoder:   decoderName,
		Data: map[string]interface{}{
			"tx_hash":       txHash,
			"decode_time":   decodeTime,
			"actions_count": len(actions),
			"success":       err == nil,
		},
		Error: err,
	}

	// Send notification non-blockingly
	select {
	case da.notification <- event:
	default:
		// Channel is full, log this but don't block
		log.Printf("Audit notification channel full, dropping event for tx %s", txHash)
	}

	return actions, err
}

// updateStats updates the statistics for a decoder
// Following code quality and performance tracking recommendations
func (da *DecoderAuditor) updateStats(decoderName string, success bool, decodeTime time.Duration, actions []types.DecodedAction) {
	da.mu.Lock()
	defer da.mu.Unlock()

	stats, exists := da.stats[decoderName]
	if !exists {
		stats = &DecoderStats{
			TotalDecoded:   0,
			TotalFailed:    0,
			AvgDecodeTime:  0,
			MaxDecodeTime:  0,
			MinDecodeTime:  time.Duration(1<<63 - 1), // Max possible duration
			SuccessRate:    0,
			TotalLiquidity: big.NewInt(0), // Initialize as big.Int to prevent overflow
			ErrorDetails:   make([]string, 0),
		}
		da.stats[decoderName] = stats
	}

	if success {
		stats.TotalDecoded++
	} else {
		stats.TotalFailed++
		// Add error details for better tracking
		stats.ErrorDetails = append(stats.ErrorDetails, fmt.Sprintf("Error at: %v", time.Now()))
	}

	// Update decode time statistics
	stats.LastDecodeTime = decodeTime

	// Update performance metrics
	if decodeTime > stats.MaxDecodeTime {
		stats.MaxDecodeTime = decodeTime
	}
	if decodeTime < stats.MinDecodeTime {
		stats.MinDecodeTime = decodeTime
	}

	// Update overall performance metrics
	if decodeTime > da.performanceMetrics.MaxDecodeTime {
		da.performanceMetrics.MaxDecodeTime = decodeTime
	}
	if decodeTime < da.performanceMetrics.MinDecodeTime {
		da.performanceMetrics.MinDecodeTime = decodeTime
	}

	totalOperations := stats.TotalDecoded + stats.TotalFailed
	if totalOperations > 1 {
		// Use a more precise calculation for average
		stats.AvgDecodeTime = time.Duration((int64(stats.AvgDecodeTime)*int64(totalOperations-1) + int64(decodeTime)) / int64(totalOperations))
	} else {
		stats.AvgDecodeTime = decodeTime
	}

	// Calculate success rate
	if totalOperations > 0 {
		stats.SuccessRate = float64(stats.TotalDecoded) / float64(totalOperations)
	}

	// Estimate liquidity processed from actions (using big.Int to prevent overflow)
	for _, action := range actions {
		if action.AmountIn != nil && action.AmountIn.Sign() > 0 {
			stats.TotalLiquidity = new(big.Int).Add(stats.TotalLiquidity, action.AmountIn)
		}
		if action.AmountOut != nil && action.AmountOut.Sign() > 0 {
			stats.TotalLiquidity = new(big.Int).Add(stats.TotalLiquidity, action.AmountOut)
		}
	}

	stats.LastUpdated = time.Now()

	// Update performance metrics tracking
	da.performanceMetrics.TotalDecodeTime += decodeTime
	da.performanceMetrics.DecodeCount++
}

// GetStats returns the current statistics for all decoders
// Following code quality and security recommendations
func (da *DecoderAuditor) GetStats() map[string]*DecoderStats {
	da.mu.RLock()
	defer da.mu.RUnlock()

	// Return a copy of the stats to prevent race conditions
	result := make(map[string]*DecoderStats)
	for name, stats := range da.stats {
		// Create a copy of the error details slice
		errorDetails := make([]string, len(stats.ErrorDetails))
		copy(errorDetails, stats.ErrorDetails)

		result[name] = &DecoderStats{
			TotalDecoded:   stats.TotalDecoded,
			TotalFailed:    stats.TotalFailed,
			AvgDecodeTime:  stats.AvgDecodeTime,
			MaxDecodeTime:  stats.MaxDecodeTime,
			MinDecodeTime:  stats.MinDecodeTime,
			LastDecodeTime: stats.LastDecodeTime,
			SuccessRate:    stats.SuccessRate,
			LastUpdated:    stats.LastUpdated,
			TotalLiquidity: new(big.Int).Set(stats.TotalLiquidity), // Deep copy to prevent external modification
			ErrorDetails:   errorDetails,
		}
	}
	return result
}

// GetPerformanceMetrics returns performance metrics for monitoring
// Following the performance monitoring recommendations
func (da *DecoderAuditor) GetPerformanceMetrics() *DecoderPerformanceMetrics {
	da.mu.RLock()
	defer da.mu.RUnlock()

	if da.performanceMetrics.DecodeCount == 0 {
		return &DecoderPerformanceMetrics{
			MaxDecodeTime:   0,
			MinDecodeTime:   0,
			TotalDecodeTime: 0,
			DecodeCount:     0,
		}
	}

	return &DecoderPerformanceMetrics{
		MaxDecodeTime:   da.performanceMetrics.MaxDecodeTime,
		MinDecodeTime:   da.performanceMetrics.MinDecodeTime,
		TotalDecodeTime: da.performanceMetrics.TotalDecodeTime,
		DecodeCount:     da.performanceMetrics.DecodeCount,
	}
}

// reportStats periodically reports audit statistics
// Following the audit framework's recommendation for comprehensive monitoring
func (da *DecoderAuditor) reportStats(ctx context.Context) {
	ticker := time.NewTicker(da.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := da.GetStats()
			perfMetrics := da.GetPerformanceMetrics()

			// Log overall performance metrics
			log.Printf("Decoder Auditor Performance Metrics: TotalDecodeTime=%v, Count=%d, AvgTime=%v, MaxTime=%v, MinTime=%v",
				perfMetrics.TotalDecodeTime, perfMetrics.DecodeCount,
				time.Duration(int64(perfMetrics.TotalDecodeTime)/perfMetrics.DecodeCount),
				perfMetrics.MaxDecodeTime, perfMetrics.MinDecodeTime)

			for decoderName, stat := range stats {
				log.Printf("Decoder Auditor Stats - %s: Total=%d, Failed=%d, SuccessRate=%.2f%%, AvgTime=%v, MaxTime=%v, MinTime=%v, Liquidity=%d, Errors=%d",
					decoderName, stat.TotalDecoded, stat.TotalFailed, stat.SuccessRate*100, stat.AvgDecodeTime,
					stat.MaxDecodeTime, stat.MinDecodeTime, stat.TotalLiquidity, len(stat.ErrorDetails))

				// Check for alert conditions based on audit framework
				if da.config.EnableAlerts {
					alertTriggered := false

					if stat.SuccessRate < da.config.MinSuccessRate {
						alertEvent := AuditEvent{
							Timestamp: time.Now(),
							EventType: "alert_low_success_rate",
							Decoder:   decoderName,
							Message:   fmt.Sprintf("Decoder success rate below threshold: %.2f%% < %.2f%%", stat.SuccessRate*100, da.config.MinSuccessRate*100),
							Data: map[string]interface{}{
								"current_rate": stat.SuccessRate,
								"threshold":    da.config.MinSuccessRate,
								"decoder_name": decoderName,
							},
						}
						select {
						case da.notification <- alertEvent:
							alertTriggered = true
						default:
							log.Println("Alert notification channel full")
						}
					}

					if stat.AvgDecodeTime > da.config.MaxDecodeTime {
						alertEvent := AuditEvent{
							Timestamp: time.Now(),
							EventType: "alert_high_decode_time",
							Decoder:   decoderName,
							Message:   fmt.Sprintf("Decoder average decode time above threshold: %v > %v", stat.AvgDecodeTime, da.config.MaxDecodeTime),
							Data: map[string]interface{}{
								"current_time": stat.AvgDecodeTime,
								"threshold":    da.config.MaxDecodeTime,
								"decoder_name": decoderName,
							},
						}
						select {
						case da.notification <- alertEvent:
							alertTriggered = true
						default:
							log.Println("Alert notification channel full")
						}
					}

					// Trigger a general alert event if any alert was generated
					if alertTriggered {
						// Update alert metrics in the audit manager if available
						// This would typically be done through a callback or shared metrics system
					}
				}
			}
		}
	}
}

// handleNotifications processes audit notifications
func (da *DecoderAuditor) handleNotifications(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-da.notification:
			// Log the event
			log.Printf("AUDIT EVENT: %s - %s: %s", event.EventType, event.Decoder, event.Message)

			// Handle specific event types
			switch event.EventType {
			case "alert_low_success_rate", "alert_high_decode_time":
				log.Printf("ALERT: %s - %s", event.EventType, event.Message)
				// In a real system, this might trigger alerts to monitoring systems
			}
		}
	}
}

// GetNotificationChannel returns the channel for audit notifications
func (da *DecoderAuditor) GetNotificationChannel() <-chan AuditEvent {
	return da.notification
}
