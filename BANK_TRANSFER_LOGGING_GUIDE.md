# Bank Transfer Comprehensive Logging Implementation

## Logging Coverage Added

### All Bank Transfer Operations Now Log:

#### 1. **initiateWalletTransfer** (Transfer OUT Initiation)
- ✅ Transfer request received (endpoint, method)
- ✅ Request decoding and validation
- ✅ User lookup (phone, user_id, current balance)
- ✅ Balance sufficiency check
- ✅ Duplicate request detection
- ✅ Transfer record creation
- ✅ Balance reservation
- ✅ Success/failure at each step

**Log Fields Captured:**
- transaction_id, user_id, phone_number, amount
- bank_account, account_holder, bank_code, bank_reference
- available balance, reserved amount, new balance
- operation status and timestamps

#### 2. **verifyWalletTransfer** (Transfer OUT Completion)
- ✅ Verification request received
- ✅ Transaction lookup
- ✅ Status update to completed
- ✅ Success response

**Log Fields Captured:**
- bank_reference, status, user_id, amount
- response code and transaction details

#### 3. **getWalletTransferStatus** (Transfer OUT Status Query)
- ✅ Status query request
- ✅ Transaction lookup (found/not found)
- ✅ Complete status details returned
- ✅ Response sent

**Log Fields Captured:**
- transaction_id, status, amount, phone_number
- bank_account, bank_reference, failure_reason
- created_at, completed_at timestamps

#### 4. **handleWalletTransferFailure** (Transfer OUT Failure & Refund)
- ✅ Failure notification received
- ✅ Transaction lookup
- ✅ Status update to failed
- ✅ Balance refund processing
- ✅ Refund success/skip with reasons
- ✅ Success response

**Log Fields Captured:**
- bank_reference, failure_reason, user_id, amount
- refund_amount, previous_status, new_status
- refund_processed flag, response code

#### 5. **loadWallet** (Transfer IN Initiation)
- ✅ Request received (phone_number, amount, bank_reference)
- ✅ User lookup
- ✅ Duplicate detection
- ✅ Balance update
- ✅ Success/failure tracking

#### 6. **verifyWalletLoad** (Transfer IN Completion)
- ✅ Verification request received
- ✅ Transaction lookup
- ✅ Balance update
- ✅ Status update to completed
- ✅ Success response

#### 7. **getWalletLoadStatus** (Transfer IN Status Query)
- ✅ Status query request
- ✅ Transaction found/not found
- ✅ Complete status details returned

#### 8. **handleWalletLoadFailure** (Transfer IN Failure)
- ✅ Failure notification received
- ✅ Transaction status update
- ✅ Failure reason stored
- ✅ Success response

## Log Level Distribution

| Level | Usage | Count |
|-------|-------|-------|
| **DEBUG** | Request details, internal state | ~15+ |
| **INFO** | Major operation steps, success | ~35+ |
| **WARN** | Validation failures, insufficient balance | ~20+ |
| **ERROR** | Database errors, exceptions | ~15+ |

**Total Logging Points:** 85+ structured log entries

## Log Output Format (JSON)

All logs output as JSON with structure:
```json
{
  "level": "info/warn/error/debug",
  "msg": "Human-readable message",
  "time": "2026-05-17 14:30:45",
  "field1": "value1",
  "field2": 123,
  "field3": true
}
```

## Sample Log Trace (Success Path)

```json
{"level":"info","msg":"Transfer request received","endpoint":"/bank-api/v1/wallet/transfer","method":"POST","time":"..."}
{"level":"debug","msg":"Request decoded","phone_number":"+977-9800000000","amount":5000,"bank_reference":"TXN-001","bank_account":"1234567890","time":"..."}
{"level":"info","msg":"Validating user","phone_number":"+977-9800000000","time":"..."}
{"level":"info","msg":"User found","user_id":"user-123","balance":10000,"time":"..."}
{"level":"info","msg":"Creating transfer record","transaction_id":"abc123","user_id":"user-123","amount":5000,"bank_account":"1234567890","bank_code":"IME","bank_reference":"TXN-001","time":"..."}
{"level":"info","msg":"Transfer record created","transaction_id":"abc123","status":"pending","time":"..."}
{"level":"info","msg":"Reserving balance","user_id":"user-123","amount":5000,"current_balance":10000,"time":"..."}
{"level":"info","msg":"Balance reserved successfully","user_id":"user-123","reserved":5000,"new_balance":5000,"transaction_id":"abc123","time":"..."}
{"level":"info","msg":"Wallet transfer initiated successfully","transaction_id":"abc123","phone_number":"+977-9800000000","amount":5000,"bank_account":"1234567890","account_holder":"John Doe","bank_code":"IME","bank_reference":"TXN-001","time":"..."}
```

## Sample Log Trace (Error Path - Insufficient Balance)

```json
{"level":"info","msg":"Transfer request received","endpoint":"/bank-api/v1/wallet/transfer","method":"POST","time":"..."}
{"level":"debug","msg":"Request decoded","phone_number":"+977-9800000000","amount":5000,"bank_reference":"TXN-001","bank_account":"1234567890","time":"..."}
{"level":"info","msg":"Validating user","phone_number":"+977-9800000000","time":"..."}
{"level":"info","msg":"User found","user_id":"user-123","balance":2000,"time":"..."}
{"level":"warn","msg":"Insufficient balance","user_id":"user-123","available":2000,"requested":5000,"shortfall":3000,"code":"BALANCE_001","time":"..."}
```

## Sample Log Trace (Failure Path - With Refund)

```json
{"level":"info","msg":"Failure notification received","bank_reference":"TXN-001","failure_reason":"Account closed","time":"..."}
{"level":"info","msg":"Transaction lookup","bank_reference":"TXN-001","time":"..."}
{"level":"debug","msg":"Transaction found","bank_reference":"TXN-001","user_id":"user-123","amount":5000,"current_status":"pending","reason":"Account closed","time":"..."}
{"level":"info","msg":"Refund processing started","bank_reference":"TXN-001","user_id":"user-123","amount":5000,"previous_status":"pending","time":"..."}
{"level":"info","msg":"Balance refunded successfully","user_id":"user-123","refund_amount":5000,"new_balance":10000,"time":"..."}
{"level":"info","msg":"Transfer failure recorded and refunded","bank_reference":"TXN-001","user_id":"user-123","failure_reason":"Account closed","amount":5000,"refund_processed":true,"response_code":"200","time":"..."}
```

## Monitoring & Alerts Based on Logs

### Error Monitoring
```
grep '"level":"error"' logs.json | jq '.msg, .code, .error'
```

### Warning Monitoring  
```
grep '"level":"warn"' logs.json | jq '.msg, .code'
```

### Transfer Success Rate
```
grep '"msg":"Wallet transfer initiated successfully"' logs.json | wc -l
```

### Failed Transfers (Refunded)
```
grep '"msg":"Transfer failure recorded and refunded"' logs.json | wc -l
```

### Average Refund Amount
```
grep '"msg":"Balance refunded successfully"' logs.json | jq '.refund_amount' | awk '{sum+=$1} END {print sum/NR}'
```

### Insufficient Balance Attempts
```
grep '"msg":"Insufficient balance"' logs.json | jq '.shortfall' | awk '{sum+=$1} END {print sum}'
```

## Testing Log Output

### View Real-Time Logs (Docker)
```bash
docker logs -f payment-system-bank-api-1 | grep '{"level'
```

### Pretty-Print Logs (using jq if installed)
```bash
docker logs payment-system-bank-api-1 | jq .
```

### Filter Errors Only
```bash
docker logs payment-system-bank-api-1 | grep '"level":"error"'
```

### Filter Warnings Only
```bash
docker logs payment-system-bank-api-1 | grep '"level":"warn"'
```

### Filter Specific Transaction
```bash
docker logs payment-system-bank-api-1 | grep 'TXN-001'
```

### Follow Transfer Completion Flow
```bash
docker logs payment-system-bank-api-1 | grep -E '"transaction_id":"abc123"'
```

## Log Storage & Aggregation

For production environments, pipe logs to:

### Option 1: ELK Stack (Elasticsearch, Logstash, Kibana)
```
Bank API Logs → Filebeat → Logstash → Elasticsearch → Kibana
```

### Option 2: Datadog
```
Bank API Logs → Datadog Agent → Datadog Dashboard
```

### Option 3: Cloud Logging (Google Cloud, AWS CloudWatch)
```
Bank API Logs → Cloud Logging API → Cloud Storage
```

### Option 4: Local File
```bash
docker logs payment-system-bank-api-1 > logs/bank-api.json
```

## Log Analysis Queries

### All transfers from a specific user
```
cat logs/bank-api.json | jq '.user_id == "user-123"'
```

### All failed transfers
```
cat logs/bank-api.json | jq 'select(.msg | contains("failed"))'
```

### Transfers on a specific date
```
cat logs/bank-api.json | jq 'select(.time | contains("2026-05-17"))'
```

### Total amount transferred
```
cat logs/bank-api.json | jq 'select(.msg | contains("Wallet transfer initiated successfully")) | .amount' | awk '{sum+=$1} END {print sum}'
```

## Audit Trail Capabilities

With comprehensive logging, you can now:

1. **Track Every Transfer**
   - Who: phone_number, user_id
   - What: amount, bank_account
   - When: created_at, completed_at timestamps
   - Status: success, failure, refunded

2. **Debug Issues**
   - Balance validation failures
   - Database errors
   - User lookup failures
   - Duplicate request handling

3. **Monitor System Health**
   - Transfer success rate
   - Refund frequency
   - Database error patterns
   - API response times

4. **Compliance & Security**
   - Complete audit trail of all transactions
   - Failure reasons and refund tracking
   - Error categorization and codes
   - Timestamp of every operation

5. **Performance Analysis**
   - Measure operation duration between log points
   - Identify bottlenecks
   - Track concurrent operations
