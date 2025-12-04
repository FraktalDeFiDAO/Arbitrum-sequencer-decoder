# Auditing Agents

The Arbitrum Sequencer Decoder includes a comprehensive auditing system to monitor and validate the various components of the arbitrage system. This system helps ensure the reliability, accuracy, and profitability of the arbitrage operations.

## Components

### 1. Decoder Auditor

The Decoder Auditor monitors the performance and accuracy of the DEX transaction decoders.

#### Features
- Tracks decoding success rates
- Measures average, min, and max decode times
- Monitors total liquidity processed
- Provides configuration thresholds for alerts
- Reports statistics at regular intervals
- Tracks error details for better debugging
- Implements performance metrics tracking

#### Configuration
- `Interval`: How often to report statistics (default: 30s)
- `EnableStats`: Whether to enable statistics collection (default: true)
- `EnableAlerts`: Whether to enable alerting (default: true)
- `MinSuccessRate`: Minimum success rate before alerting (default: 95%)
- `MaxDecodeTime`: Maximum average decode time before alerting (default: 100ms)

### 2. Arbitrage Auditor

The Arbitrage Auditor validates arbitrage opportunities to ensure they are profitable and low-risk.

#### Features
- Validates opportunity profitability
- Calculates risk scores based on probability and inherent risk factors
- Caches validation results
- Monitors for expired opportunities
- Provides validation before execution
- Implements security metrics tracking
- Tracks performance metrics for validation operations

### 3. System Health Auditor

The System Health Auditor monitors the overall health of all system components.

#### Features
- Component status tracking (healthy/degraded/unhealthy)
- Response time monitoring
- Error detection and reporting
- Comprehensive system statistics
- Health check registration for custom components
- Implements security and performance metrics
- Maintains audit trails for compliance

### 4. Audit Manager

The Audit Manager coordinates all auditing activities and provides a unified interface.

#### Features
- Centralized control of all auditors
- Unified statistics access
- Easy integration with existing components
- Thread-safe operation
- Aggregated audit metrics
- Implements audit framework compliance

## Extended Capabilities

### Performance Monitoring
- All auditors now track comprehensive performance metrics
- Min/Max/Average time calculations
- Performance trend analysis
- Resource usage monitoring

### Security Tracking
- Risk event detection and monitoring
- Invalid opportunities tracking
- Suspicious activity monitoring
- Security compliance metrics

### Audit Trail and Compliance
- Complete audit trail logging
- Compliance with audit framework requirements
- Detailed event tracking
- Historical analysis capabilities

## Integration

To integrate the auditing system into your application:

1. Create an AuditManager instance:
   ```go
   auditConfig := &AuditConfig{
       Interval: 30 * time.Second,
       EnableStats: true,
       EnableAlerts: true,
       MinSuccessRate: 0.90,
       MaxDecodeTime: 200 * time.Millisecond,
   }

   auditManager := NewAuditManager(auditConfig)
   ```

2. Initialize auditors with system components:
   ```go
   auditManager.InitializeAuditors(decoder, oracle, arbEngine, executor)
   ```

3. Start the auditing system:
   ```go
   go auditManager.Start(ctx)
   ```

4. Use audited operations:
   ```go
   // Decode with auditing
   actions, err := auditManager.AuditTransaction(txHash, func() ([]types.DecodedAction, error) {
       return decoder.Decode(tx, toAddress)
   })

   // Audit an arbitrage opportunity
   validation, err := auditManager.AuditOpportunity(opportunity)

   // Validate and execute with auditing
   err := auditManager.ValidateAndExecute(executor, opportunity)
   ```

5. Access comprehensive metrics:
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

## Statistics and Monitoring

Each auditor provides access to extensive statistics:

- **Decoder Stats**: Success rates, decode times (avg/min/max), liquidity processed, error details
- **Arbitrage Stats**: Validated opportunities, profitability, risk scores
- **System Health Stats**: Uptime, active components, transaction volume
- **Component Status**: Individual component health status
- **Performance Metrics**: Detailed performance tracking per system component
- **Security Metrics**: Risk events, invalid opportunities, suspicious activities

## Alerts and Notifications

The auditing system provides both internal monitoring and external notification capabilities through channels that can be monitored for audit events. Critical issues trigger alerts that can be integrated with external monitoring systems.

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