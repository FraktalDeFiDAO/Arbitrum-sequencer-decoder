# Auditing Agents Documentation

## Overview

The Arbitrum Sequencer Decoder includes comprehensive auditing agents to monitor and validate the various components of the arbitrage system. These agents ensure reliability, accuracy, performance, and security of the arbitrage operations, following the project's audit framework standards.

## Components

### 1. Audit Manager

The AuditManager coordinates all auditing activities and provides a unified interface for the entire auditing system.

#### Key Features
- Centralized control of all auditors
- Unified statistics access
- Thread-safe operations
- Aggregated audit metrics
- Audit framework compliance

#### Usage
```go
auditManager := NewAuditManager(config)
auditManager.InitializeAuditors(decoder, oracle, arbEngine, executor)
go auditManager.Start(ctx)

// Perform audited operations
actions, err := auditManager.AuditTransaction(txHash, decodeFunc)
validation, err := auditManager.AuditOpportunity(opportunity)
```

### 2. Decoder Auditor

The DecoderAuditor monitors the performance and accuracy of the DEX transaction decoders.

#### Key Features
- Tracks decoding success rates
- Measures average, min, and max decode times
- Monitors total liquidity processed
- Provides configuration thresholds for alerts
- Tracks error details for better debugging
- Implements performance metrics tracking

#### Metrics Tracked
- Success rates and failure counts
- Decode time statistics (average, min, max)
- Total liquidity processed
- Error details and timestamps

### 3. Arbitrage Auditor

The ArbitrageAuditor validates arbitrage opportunities to ensure they are profitable and low-risk.

#### Key Features
- Validates opportunity profitability
- Calculates risk scores based on probability and inherent risk factors
- Caches validation results
- Monitors for expired opportunities
- Provides validation before execution
- Implements security metrics tracking
- Tracks performance metrics for validation operations

#### Security Features
- Risk event detection
- Invalid opportunity tracking
- Suspicious activity monitoring

### 4. System Health Auditor

The SystemHealthAuditor monitors the overall health of all system components.

#### Key Features
- Component status tracking (healthy/degraded/unhealthy)
- Response time monitoring
- Error detection and reporting
- Comprehensive system statistics
- Health check registration for custom components
- Implements security and performance metrics
- Maintains audit trails for compliance

## Integration Patterns

### Standard Integration
```go
// Initialize audit manager with configuration
auditConfig := &AuditConfig{
    Interval:       30 * time.Second,
    EnableStats:    true,
    EnableAlerts:   true,
    MinSuccessRate: 0.90,
    MaxDecodeTime:  200 * time.Millisecond,
}

auditManager := NewAuditManager(auditConfig)
auditManager.InitializeAuditors(decoder, oracle, arbEngine, executor)

// Start auditing system
go auditManager.Start(ctx)

// Use audited operations
actions, err := auditManager.AuditTransaction(txHash, decodeFunc)
validation, err := auditManager.AuditOpportunity(opportunity)
```

### Metrics Collection
```go
// Get aggregated audit metrics
metrics := auditManager.GetAuditMetrics()

// Get specific auditor metrics
decoderStats := auditManager.GetDecoderStats()
arbitrageStats := auditManager.GetArbitrageStats()
healthStats := auditManager.GetSystemHealthStats()

// Get security and performance metrics
decoderAuditor := /* reference to decoder auditor */
perfMetrics := decoderAuditor.GetPerformanceMetrics()
```

## Best Practices

1. Always initialize auditing before starting the main application logic
2. Configure appropriate thresholds for your specific use case
3. Monitor audit statistics regularly to identify performance issues
4. Use the audit system to detect and respond to failed operations
5. Implement proper logging to track audit events
6. Ensure compliance with audit framework requirements
7. Regularly review audit trails and security metrics
8. Set up alerts for critical security and performance indicators
9. Use performance metrics to optimize system efficiency
10. Maintain audit compliance for regulatory requirements

## Security Considerations

1. All auditors implement proper security metrics tracking
2. Risk events are detected and logged
3. Invalid operations are tracked and reported
4. Audit trails maintain compliance with audit framework
5. Sensitive information is not exposed in logs

## Performance Monitoring

1. All operations are timed and tracked
2. Performance metrics are aggregated
3. Min/Max/Average calculations are maintained
4. Resource usage is monitored
5. Performance trends are analyzed

## Error Handling

1. All errors are properly logged
2. Error details are stored for debugging
3. Error counts are aggregated
4. Error patterns are analyzed for improvements