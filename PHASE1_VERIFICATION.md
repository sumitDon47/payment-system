# Phase 1 Verification Guide

## Quick Status Check

### 1. Verify Services are Running
```bash
cd c:\Users\Raja\Desktop\payment-system

# Check all containers
docker compose ps

# Expected output:
# SERVICE           STATUS
# payment-service   Up X minutes
# user-service      Up X minutes  
# postgres          Up X minutes (healthy)
# redis             Up X minutes (healthy)
# zookeeper         Up X minutes (healthy)
```

### 2. Check Structured Logs (JSON Format)
```bash
# View recent payment service logs with structured JSON
docker logs payment-system-payment-service-1 --tail 20

# Expected output includes:
# {"level":"info","msg":"Payment Service starting",...}
# {"level":"info","msg":"Database connected",...}
# {"level":"info","msg":"Prometheus metrics server started",...}
```

### 3. Verify User Service is Accessible
```bash
# Test user service HTTP endpoint
Invoke-WebRequest -Uri http://localhost:8082/user/1 -UseBasicParsing

# Should return user data (or 404 if user doesn't exist)
```

### 4. Check Prometheus Metrics (Inside Docker Network)
```bash
# From inside payment-service container
docker exec payment-system-payment-service-1 wget -O - http://localhost:8081/metrics

# Should show metrics like:
# payment_duration_bucket
# payment_total
# error_counter_total
```

## Viewing Structured Logs

### All Payment Service Logs
```bash
docker logs payment-system-payment-service-1
```

### Only JSON logs (filter)
```bash
docker logs payment-system-payment-service-1 | Select-String '{"level'
```

### Follow logs in real-time
```bash
docker logs -f payment-system-payment-service-1
```

### Pretty-print logs (optional - requires jq tool)
```bash
docker logs payment-system-payment-service-1 | Select-String '{"level' | ConvertFrom-Json | Format-Table
```

## Log Format Reference

### Structured Log Fields
Every JSON log contains:
- `level` - info, warn, error, debug
- `msg` - human-readable message
- `time` - ISO 8601 timestamp
- Custom fields (context-specific)

### Example Log Entry
```json
{
  "level": "info",
  "msg": "Payment Service starting",
  "service": "payment-service",
  "time": "2026-05-08 04:16:11"
}
```

### Expected Logs at Startup
```json
{"environment":"","level":"info","msg":"Payment Service starting","service":"payment-service"}
{"level":"info","msg":"Database connected"}
{"kafka_broker":"kafka:9092","level":"info","msg":"Outbox dispatcher started"}
{"level":"info","msg":"Payment Service listening","port":"9090","protocol":"gRPC"}
{"level":"info","msg":"Prometheus metrics server started","port":"8081","path":"/metrics"}
```

## Available Metrics

### Payment Counter
```
payment_total{currency="NPR",status="success"} N
payment_total{currency="NPR",status="failed"} M
```

### Payment Duration (in seconds)
```
payment_duration_bucket{le="0.005",operation="send_payment"} X
payment_duration_bucket{le="0.01",operation="send_payment"} Y
```

### Error Counter
```
error_counter_total{error_type="insufficient_funds"} N
error_counter_total{error_type="account_locked"} M
```

## Testing the Payment Flow

### 1. Send a Payment (requires gRPC client)
```bash
# Using grpcurl (if available)
grpcurl -plaintext \
  -d '{"sender_id":"user-1","receiver_id":"user-2","amount":1000,"currency":"NPR"}' \
  localhost:9090 payment.PaymentService/SendPayment
```

### 2. Check Logs for Structured Output
```bash
docker logs payment-system-payment-service-1 --tail 30 | Select-String "sender_id|receiver_id|payment_total"
```

### 3. Verify Metrics Incremented
```bash
docker exec payment-system-payment-service-1 wget -O - http://localhost:8081/metrics | grep "payment_total"
```

## Troubleshooting

### Service Won't Start
```bash
# Check logs
docker logs payment-system-payment-service-1

# Common issues:
# 1. Database not healthy - wait longer, check postgres logs
# 2. Kafka not ready - ensure zookeeper started first
# 3. Port conflicts - check if :9090 is available
```

### No JSON Logs Appearing
```bash
# Verify log level is set
docker exec payment-system-payment-service-1 echo $LOG_LEVEL

# Try setting it explicitly
docker stop payment-system-payment-service-1
docker start payment-system-payment-service-1
```

### Metrics Not Available
```bash
# Check if metrics server started (should be in logs)
docker logs payment-system-payment-service-1 | grep "metrics server"

# If missing, the binary may have crashed - check full logs
```

## Next Verification Steps

After confirming everything runs:

1. **Send a test payment** to generate logs with business context
2. **Monitor logs** and watch structured fields appear
3. **Check metrics** to see counters incrementing
4. **Verify error handling** by trying invalid payments

---

**Quick Start Summary**:
1. `docker compose ps` - Verify services
2. `docker logs payment-system-payment-service-1` - Check structured logs
3. `Invoke-WebRequest http://localhost:8082/user/1` - Test user service
4. Ready for Phase 2 (Grafana dashboards, distributed tracing)
