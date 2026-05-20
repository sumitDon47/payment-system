# 🚀 DEPLOYMENT, COMPLIANCE & SCALING GUIDE

**For**: Production-Grade Digital Wallet System  
**Scale Target**: 1M+ DAU, 10K+ TPS  
**Compliance**: PCI DSS, KYC/AML, Data Protection

---

## PART 1: DEVOPS & INFRASTRUCTURE

### 1.1 Kubernetes Deployment Architecture

```yaml
# kubernetes/01-namespace-rbac.yaml

apiVersion: v1
kind: Namespace
metadata:
  name: payment-system
  labels:
    name: payment-system
    monitoring: enabled

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: payment-system-sa
  namespace: payment-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: payment-system-role
rules:
- apiGroups: [""]
  resources: ["pods", "pods/logs"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: payment-system-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: payment-system-role
subjects:
- kind: ServiceAccount
  name: payment-system-sa
  namespace: payment-system

---
# Network Policy: Isolate payment-system namespace
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: payment-system-network-policy
  namespace: payment-system
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: payment-system
  - to:
    - podSelector:
        matchLabels:
          app: postgres
  - to:
    - podSelector:
        matchLabels:
          app: redis
  - to:
    - podSelector:
        matchLabels:
          app: kafka
  - ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

### 1.2 Stateful Service Deployment (PostgreSQL)

```yaml
# kubernetes/02-postgresql.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
  namespace: payment-system
data:
  max_connections: "200"
  shared_buffers: "256MB"
  effective_cache_size: "1GB"
  maintenance_work_mem: "64MB"
  checkpoint_completion_target: "0.9"
  wal_buffers: "16MB"
  default_statistics_target: "100"
  random_page_cost: "1.1"
  effective_io_concurrency: "200"
  work_mem: "1310kB"
  min_wal_size: "1GB"
  max_wal_size: "4GB"

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: payment-system
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 500Gi
  storageClassName: fast-ssd

---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: payment-system
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      securityContext:
        fsGroup: 999
      containers:
      - name: postgres
        image: postgres:15-alpine
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        - name: POSTGRES_USER
          value: payment_user
        - name: POSTGRES_DB
          value: payment_db
        
        volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
        - name: config
          mountPath: /etc/postgresql
          
        resources:
          requests:
            memory: "4Gi"
            cpu: "2"
          limits:
            memory: "8Gi"
            cpu: "4"
        
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pg_isready -U payment_user
          initialDelaySeconds: 30
          periodSeconds: 10
      
      volumes:
      - name: config
        configMap:
          name: postgres-config
  
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 500Gi

---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: payment-system
spec:
  ports:
  - port: 5432
    targetPort: 5432
  clusterIP: None
  selector:
    app: postgres
```

### 1.3 Redis Cluster Deployment

```yaml
# kubernetes/03-redis-cluster.yaml

apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: payment-system
spec:
  serviceName: redis
  replicas: 3
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command:
        - redis-server
        - /usr/local/etc/redis/redis.conf
        - --cluster-enabled
        - "yes"
        - --cluster-config-file
        - /data/nodes.conf
        - --cluster-node-timeout
        - "5000"
        - --appendonly
        - "yes"
        
        ports:
        - containerPort: 6379
          name: client
        - containerPort: 16379
          name: gossip
        
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /usr/local/etc/redis
        
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      
      volumes:
      - name: config
        configMap:
          name: redis-config
  
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 100Gi

---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: payment-system
spec:
  ports:
  - port: 6379
    name: client
  - port: 16379
    name: gossip
  clusterIP: None
  selector:
    app: redis
```

### 1.4 Kafka Cluster

```yaml
# kubernetes/04-kafka.yaml

apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: kafka
  namespace: payment-system
spec:
  serviceName: kafka
  replicas: 3
  selector:
    matchLabels:
      app: kafka
  template:
    metadata:
      labels:
        app: kafka
    spec:
      containers:
      - name: kafka
        image: confluentinc/cp-kafka:7.5.0
        ports:
        - containerPort: 9092
          name: broker
        - containerPort: 9999
          name: metrics
        
        env:
        - name: KAFKA_BROKER_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: KAFKA_ZOOKEEPER_CONNECT
          value: zookeeper:2181
        - name: KAFKA_ADVERTISED_LISTENERS
          value: PLAINTEXT://kafka-$(KAFKA_BROKER_ID).kafka:9092
        - name: KAFKA_LISTENERS
          value: PLAINTEXT://0.0.0.0:9092
        - name: KAFKA_INTER_BROKER_LISTENER_NAME
          value: PLAINTEXT
        - name: KAFKA_AUTO_CREATE_TOPICS_ENABLE
          value: "true"
        - name: KAFKA_LOG_RETENTION_HOURS
          value: "168"
        - name: KAFKA_NUM_NETWORK_THREADS
          value: "8"
        - name: KAFKA_NUM_IO_THREADS
          value: "8"
        - name: KAFKA_SOCKET_SEND_BUFFER_BYTES
          value: "102400"
        - name: KAFKA_SOCKET_RECEIVE_BUFFER_BYTES
          value: "102400"
        - name: KAFKA_SOCKET_REQUEST_MAX_BYTES
          value: "104857600"
        
        volumeMounts:
        - name: data
          mountPath: /var/lib/kafka/data
        
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
  
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 200Gi

---
apiVersion: v1
kind: Service
metadata:
  name: kafka
  namespace: payment-system
spec:
  ports:
  - port: 9092
    name: broker
  - port: 9999
    name: metrics
  clusterIP: None
  selector:
    app: kafka
```

---

## PART 2: MONITORING & OBSERVABILITY

### 2.1 Prometheus Setup

```yaml
# kubernetes/monitoring/prometheus.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    
    alerting:
      alertmanagers:
      - static_configs:
        - targets:
          - alertmanager:9093
    
    rule_files:
    - '/etc/prometheus/rules.yaml'
    
    scrape_configs:
    - job_name: 'kubernetes-apiservers'
      kubernetes_sd_configs:
      - role: endpoints
      
    - job_name: 'payment-system-services'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - payment-system
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
    
    - job_name: 'postgres'
      static_configs:
      - targets: ['postgres-exporter:9187']
    
    - job_name: 'redis'
      static_configs:
      - targets: ['redis-exporter:9121']
    
    - job_name: 'kafka'
      static_configs:
      - targets: ['kafka-exporter:9308']

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      serviceAccountName: prometheus
      containers:
      - name: prometheus
        image: prom/prometheus:latest
        args:
        - '--config.file=/etc/prometheus/prometheus.yml'
        - '--storage.tsdb.path=/prometheus'
        - '--storage.tsdb.retention.time=30d'
        
        ports:
        - containerPort: 9090
        
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        - name: storage
          mountPath: /prometheus
        
        resources:
          requests:
            memory: "2Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "1000m"
      
      volumes:
      - name: config
        configMap:
          name: prometheus-config
      - name: storage
        emptyDir: {}

---
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  namespace: monitoring
spec:
  ports:
  - port: 9090
    targetPort: 9090
  selector:
    app: prometheus
```

### 2.2 Distributed Tracing (Jaeger)

```yaml
# kubernetes/monitoring/jaeger.yaml

apiVersion: v1
kind: Service
metadata:
  name: jaeger-collector
  namespace: monitoring
spec:
  ports:
  - name: otlp-grpc
    port: 4317
    targetPort: 4317
  - name: otlp-http
    port: 4318
    targetPort: 4318
  - name: zipkin
    port: 9411
    targetPort: 9411
  selector:
    app: jaeger

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger
  namespace: monitoring
spec:
  replicas: 2
  selector:
    matchLabels:
      app: jaeger
  template:
    metadata:
      labels:
        app: jaeger
    spec:
      containers:
      - name: jaeger
        image: jaegertracing/all-in-one:latest
        ports:
        - containerPort: 4317  # OTLP gRPC
        - containerPort: 4318  # OTLP HTTP
        - containerPort: 9411  # Zipkin
        - containerPort: 16686 # UI
        
        env:
        - name: COLLECTOR_OTLP_ENABLED
          value: "true"
        - name: SPAN_STORAGE_TYPE
          value: elasticsearch
        - name: ES_SERVER_URLS
          value: "http://elasticsearch:9200"
        
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"

---
apiVersion: v1
kind: Service
metadata:
  name: jaeger-ui
  namespace: monitoring
spec:
  ports:
  - port: 16686
    targetPort: 16686
  selector:
    app: jaeger
  type: LoadBalancer
```

---

## PART 3: COMPLIANCE & AUDIT

### 3.1 KYC/AML Implementation

```sql
-- database/kyc_aml_schema.sql

-- KYC Status Tracking
CREATE TABLE kyc_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    -- Personal Info
    full_name VARCHAR(200) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender VARCHAR(10),
    nationality VARCHAR(50),
    
    -- Address
    street_address VARCHAR(255),
    city VARCHAR(100),
    state_province VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(50),
    
    -- Document
    document_type VARCHAR(20), -- PASSPORT, DRIVERS_LICENSE, NATIONAL_ID
    document_number VARCHAR(100),
    document_issue_date DATE,
    document_expiry_date DATE,
    document_country VARCHAR(50),
    document_image_url TEXT,
    
    -- Verification
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, submitted, approved, rejected
    kyc_level INTEGER DEFAULT 1, -- 1=basic, 2=intermediate, 3=full
    verified_at TIMESTAMP,
    verified_by_admin_id UUID,
    rejection_reason TEXT,
    
    -- Timestamps
    submitted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- AML Risk Assessment
CREATE TABLE aml_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    -- Risk Scores
    transaction_risk_score NUMERIC(5, 2), -- 0-100
    pep_risk_score NUMERIC(5, 2),
    sanctions_risk_score NUMERIC(5, 2),
    behavioral_risk_score NUMERIC(5, 2),
    overall_risk_score NUMERIC(5, 2), -- Weighted average
    
    -- Flags
    is_pep BOOLEAN DEFAULT FALSE, -- Politically Exposed Person
    is_sanctioned BOOLEAN DEFAULT FALSE,
    is_high_risk_country BOOLEAN DEFAULT FALSE,
    
    -- Details
    risk_factors TEXT[], -- Array of risk factor descriptions
    recommended_action VARCHAR(50), -- NONE, MONITOR, REVIEW, SUSPEND
    
    -- Assessment
    assessed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reviewed_by_admin_id UUID,
    review_notes TEXT,
    
    CONSTRAINT overall_score_range CHECK (overall_risk_score >= 0 AND overall_risk_score <= 100)
);

-- Sanctions List Screening
CREATE TABLE sanctions_screening_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    full_name VARCHAR(200) NOT NULL,
    nationality VARCHAR(50),
    
    -- Screening Result
    matched_names TEXT[],
    confidence_score NUMERIC(5, 2),
    hit BOOLEAN DEFAULT FALSE,
    false_positive BOOLEAN DEFAULT FALSE,
    
    -- Action
    reviewed_at TIMESTAMP,
    reviewed_by_admin_id UUID,
    action_taken VARCHAR(50), -- NONE, INVESTIGATE, BLOCK
    
    -- Timestamps
    screened_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CREATE INDEX idx_sanctions_user ON sanctions_screening_log(user_id);
    CREATE INDEX idx_sanctions_screened ON sanctions_screening_log(screened_at DESC);
);

-- Large Transaction Reporting (LTR)
CREATE TABLE transaction_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    user_id UUID NOT NULL REFERENCES users(id),
    
    report_type VARCHAR(20) NOT NULL, -- LTR, STR, CTR
    amount_threshold NUMERIC(15, 2) NOT NULL,
    actual_amount NUMERIC(15, 2) NOT NULL,
    currency VARCHAR(3),
    
    -- Report Details
    report_filing_date DATE NOT NULL,
    due_date DATE NOT NULL,
    filed_with VARCHAR(50), -- FIU, GAAFIS, etc
    reference_number VARCHAR(100),
    
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, filed, acknowledged, investigated
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    filed_at TIMESTAMP,
    
    CONSTRAINT ltrlimit CHECK (amount_threshold >= 5000000) -- 5L NPR minimum for LTR
);

-- Create index for reporting
CREATE INDEX idx_ltr_filing_date ON transaction_reports(report_filing_date DESC);
CREATE INDEX idx_ltr_status ON transaction_reports(status);
```

### 3.2 Regulatory Reporting

```go
// compliance-service/internal/reporting/aml_reporter.go

package reporting

import (
    "database/sql"
    "fmt"
    "time"
)

type AMLReporter struct {
    db *sql.DB
}

// Generate Large Transaction Report (LTR)
func (ar *AMLReporter) GenerateLTR(startDate, endDate time.Time) ([]LargeTransactionReport, error) {
    // Threshold: 5,000,000 NPR
    const ltThreshold = 5000000.0
    
    rows, err := ar.db.Query(`
        SELECT 
            t.id, t.sender_id, t.amount, t.currency, t.completed_at,
            s.email as sender_email, s.full_name as sender_name,
            r.email as receiver_email, r.full_name as receiver_name
        FROM transactions t
        JOIN users s ON t.sender_id = s.id
        JOIN users r ON t.receiver_id = r.id
        WHERE t.status = 'completed' 
        AND t.amount >= $1
        AND t.completed_at BETWEEN $2 AND $3
        ORDER BY t.amount DESC
    `, ltThreshold, startDate, endDate)
    
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var reports []LargeTransactionReport
    
    for rows.Next() {
        var report LargeTransactionReport
        rows.Scan(
            &report.TransactionID, &report.SenderID, &report.Amount, &report.Currency,
            &report.TransactionDate, &report.SenderEmail, &report.SenderName,
            &report.ReceiverEmail, &report.ReceiverName,
        )
        
        reports = append(reports, report)
        
        // Insert into transaction_reports table
        ar.db.Exec(`
            INSERT INTO transaction_reports 
            (transaction_id, user_id, report_type, amount_threshold, actual_amount, 
             currency, report_filing_date, due_date, status)
            VALUES ($1, $2, 'LTR', $3, $4, $5, $6, $7, 'pending')
        `, report.TransactionID, report.SenderID, ltThreshold, report.Amount, 
           report.Currency, time.Now().Date(), time.Now().AddDate(0, 0, 7).Date())
    }
    
    return reports, nil
}

// Generate Suspicious Transaction Report (STR)
func (ar *AMLReporter) GenerateSTR(userId string, reason string) error {
    // Logic to identify suspicious patterns and create STR
    // File to FIU within 24 hours
    
    _, err := ar.db.Exec(`
        INSERT INTO transaction_reports 
        (user_id, report_type, status, report_filing_date, due_date)
        VALUES ($1, 'STR', 'pending', CURRENT_DATE, CURRENT_DATE + INTERVAL '1 day')
    `, userId)
    
    return err
}

// File report to FIU
func (ar *AMLReporter) FileReportToFIU(reportID string) error {
    // Call FIU API / Portal
    // Mark as filed
    
    _, err := ar.db.Exec(`
        UPDATE transaction_reports 
        SET status = 'filed', filed_at = NOW()
        WHERE id = $1
    `, reportID)
    
    return err
}

type LargeTransactionReport struct {
    TransactionID    string
    SenderID         string
    SenderName       string
    SenderEmail      string
    ReceiverEmail    string
    ReceiverName     string
    Amount           float64
    Currency         string
    TransactionDate  time.Time
}
```

---

## PART 4: SCALING ROADMAP

### 4.1 Scaling to 1M DAU

```
PHASE 1: Current State (10K DAU)
════════════════════════════════════
- Single PostgreSQL instance
- Single Redis instance  
- 3-node Kafka cluster
- ~3 pods per service (K8s)
- Single API gateway
- No caching layer (beyond Redis)

Metrics:
- 100 TPS capacity
- P99 latency: 500ms
- Error rate: 0.5%


PHASE 2: 100K DAU (~1000 TPS)
════════════════════════════════════
Database:
- PostgreSQL Primary + 1 Read Replica
- Connection pooling (PgBouncer)
- Replication lag monitoring
- Query optimization

Caching:
- Redis Cluster (3 master + 3 slave)
- Cache warming for frequent queries
- TTL optimization

Messaging:
- Kafka cluster expansion (5+ nodes)
- Partition rebalancing
- Consumer group scaling

Services:
- Scale to 5-7 pods per service
- Auto-scaling based on CPU/memory

Monitoring:
- Distributed tracing (Jaeger)
- Custom metrics dashboards
- SLA monitoring

Estimated Latency: P99 < 300ms
Estimated Cost: 50K USD/month


PHASE 3: 1M DAU (~10K TPS)
════════════════════════════════════
Database:
- PostgreSQL Primary + 2+ Read Replicas
- Sharding by user_id hash
- Cross-shard transactions (Saga pattern)
- Connection pooling layer
- Write-ahead log (WAL) optimization

Caching:
- Multi-tier caching (L1: Local, L2: Redis)
- Cache-aside pattern
- Probabilistic early expiration

Messaging:
- Kafka cluster (10+ nodes)
- Multiple consumer groups per topic
- Partition count = max(current throughput, expected throughput)

Services:
- 10-20 pods per service
- Horizontal pod autoscaling
- Service mesh (Istio) for traffic management

API Gateway:
- Load balancer with geographic routing
- Rate limiting per user/IP
- Request queuing during traffic spikes

Storage:
- S3 for audit logs (hot/cold archival)
- Elastic storage for backups

Monitoring:
- Real-time fraud detection dashboards
- Predictive autoscaling
- Cost optimization dashboards

Estimated Latency: P99 < 200ms
Estimated Cost: 500K USD/month


PHASE 4: 10M DAU (~100K TPS)
════════════════════════════════════
Architecture:
- Full microservices across multiple regions
- Global load balancing with failover
- Multi-region PostgreSQL with replication
- Event sourcing for audit trails
- CQRS for high-volume reads

Database:
- PostgreSQL clusters in multiple regions
- Cross-region replication
- Logical replication for selective sync
- Time-series database (InfluxDB) for metrics

Messaging:
- Kafka clusters per region
- Cross-region replication
- Schema registry for versioning

Services:
- 50-100 pods across multiple clusters
- Chaos engineering for resilience testing
- Circuit breakers and bulkheads

Performance:
- P99 latency < 100ms
- 99.99% uptime SLA
- Zero-downtime deployment

Estimated Cost: 5M USD/month
```

### 4.2 Sharding Strategy

```go
// payment-system/shared/sharding/shard_manager.go

package sharding

import (
    "fmt"
    "hash/fnv"
)

type ShardManager struct {
    shardCount int
}

// Calculate shard ID based on user ID
func (sm *ShardManager) GetShardID(userID string) int {
    h := fnv.New32a()
    h.Write([]byte(userID))
    return int(h.Sum32()) % sm.shardCount
}

// Get database connection string for shard
func (sm *ShardManager) GetShardConnection(userID string) string {
    shardID := sm.GetShardID(userID)
    host := fmt.Sprintf("postgres-shard-%d.payment-system.svc.cluster.local", shardID)
    return fmt.Sprintf(
        "postgres://user:pass@%s:5432/payment_db",
        host,
    )
}

// Example: 16 shards for 1M users = ~62.5K users per shard
// Each shard can handle 500 TPS, so 16 shards = 8K TPS total
func NewShardManager(count int) *ShardManager {
    return &ShardManager{shardCount: count}
}

// Shard assignment for P2P transfer
func (sm *ShardManager) GetTransactionShards(senderID, receiverID string) (int, int) {
    return sm.GetShardID(senderID), sm.GetShardID(receiverID)
}

// Handle cross-shard transactions using Saga pattern
type SagaCoordinator struct {
    shardManager *ShardManager
}

func (sc *SagaCoordinator) ExecuteP2PTransfer(senderID, receiverID string, amount float64) error {
    senderShard := sc.shardManager.GetShardID(senderID)
    receiverShard := sc.shardManager.GetShardID(receiverID)
    
    if senderShard == receiverShard {
        // Local transaction - simple
        return executeLocalTransaction(senderID, receiverID, amount)
    }
    
    // Distributed transaction - Saga pattern
    // Step 1: Debit sender (on sender shard)
    txn1 := executeOnShard(senderShard, `
        UPDATE user_accounts SET balance = balance - $1 WHERE user_id = $2
    `, amount, senderID)
    
    if txn1 != nil {
        return txn1
    }
    
    // Step 2: Credit receiver (on receiver shard)
    txn2 := executeOnShard(receiverShard, `
        UPDATE user_accounts SET balance = balance + $1 WHERE user_id = $2
    `, amount, receiverID)
    
    if txn2 != nil {
        // Compensating transaction: Refund sender
        executeOnShard(senderShard, `
            UPDATE user_accounts SET balance = balance + $1 WHERE user_id = $2
        `, amount, senderID)
        return txn2
    }
    
    return nil
}
```

---

## PART 5: INCIDENT RESPONSE & DISASTER RECOVERY

### 5.1 Incident Response Plan

```
SEVERITY LEVELS
═══════════════════════════════════════════

CRITICAL (P1): < 5 min response
├─ System completely down
├─ Data corruption detected
├─ Security breach detected
├─ DDoS attack
└─ Financial loss occurring

HIGH (P2): < 15 min response
├─ Payment service degraded
├─ 50%+ error rate
├─ Database replication lag > 30s
└─ Memory/CPU critically high

MEDIUM (P3): < 1 hour response
├─ Some users affected
├─ Non-critical service slow
├─ Minor security issue
└─ Performance degradation

LOW (P4): < 24 hours response
├─ Documentation issues
├─ UI/UX bugs
├─ Non-critical warnings in logs
└─ Low-severity security issues


INCIDENT RESPONSE RUNBOOK
═══════════════════════════════════════════

1. DETECTION & ALERT (5 min)
   - Prometheus alert triggers
   - PagerDuty notification
   - On-call engineer acknowledges

2. INCIDENT CLASSIFICATION (5 min)
   - Determine severity
   - Identify affected service
   - Estimate impact

3. IMMEDIATE MITIGATION (10 min)
   - For P1: Activate war room (Slack channel)
   - For P1: Page incident commander
   - For P1: Start incident recording
   - Scale up affected service pods
   - Check for circuit breaker trips
   - Review recent deployments

4. ROOT CAUSE INVESTIGATION (30 min)
   - Check logs (ELK)
   - Check metrics (Prometheus)
   - Check traces (Jaeger)
   - Review recent changes
   - Interview on-call team

5. RESOLUTION (per severity)
   - P1: 15-30 min target
   - P2: 30-60 min target
   - P3: 1-4 hours target

6. POST-INCIDENT (24 hours)
   - Root cause analysis meeting
   - Create tickets for improvements
   - Update runbooks
   - Schedule preventive measures
```

### 5.2 Disaster Recovery Plan

```yaml
# DR Strategy: RPO = 5 min, RTO = 30 min

Recovery Time Objectives (RTO):
- Database: 5 min (warm standby in another AZ)
- Cache: 2 min (rebuild from database)
- Message Queue: 10 min (replay from backup)
- Overall: 30 min to full service recovery

Recovery Point Objectives (RPO):
- Database: 5 min (continuous replication)
- Transactions: 0 min (write to primary + replica)
- Audit logs: 5 min (batch write to S3)

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: dr-procedure
  namespace: payment-system
data:
  procedure.md: |
    # DISASTER RECOVERY PROCEDURE
    
    ## Primary Failure: East Region Down
    
    1. **Detection** (Automated, < 2 min)
       - All health checks fail
       - Prometheus alerts fire
       - Failover triggers automatically
    
    2. **Failover to DR Region** (< 5 min)
       ```bash
       # Activate DR region
       kubectl apply -f dr-region-active.yaml
       
       # Verify database replication
       psql -h dr-db-primary -c "SELECT * FROM pg_stat_replication"
       
       # Promote read replica to primary
       psql -h dr-db-replica -c "SELECT pg_promote()"
       
       # Update DNS to point to DR region
       aws route53 change-resource-record-sets --hosted-zone-id Z123 \
         --change-batch file://dns-update.json
       ```
    
    3. **Restore Cache** (< 2 min)
       ```bash
       # Rebuild Redis from database
       redis-cli > FLUSHALL
       python scripts/rebuild_cache.py --source=dr-database
       ```
    
    4. **Verify Service Health** (< 3 min)
       ```bash
       # Run synthetic tests
       ./scripts/smoke-tests.sh
       
       # Monitor metrics
       watch 'kubectl top pods -n payment-system'
       ```
    
    5. **Restore Primary** (When ready)
       - Restore East region from backups
       - Sync databases
       - Conduct full testing
       - Promote East back to primary
    
    ## Data Loss Recovery
    
    - Point-in-time recovery: Available for last 30 days
    - Monthly snapshots stored in S3
    - Immutable backup storage in different region

    ## Communication Plan
    
    - Slack #incident channel
    - Customer notification (status page)
    - Internal stakeholder emails
    - Post-mortem within 24 hours
```

---

## PART 6: PRODUCTION CHECKLIST

```
PRE-DEPLOYMENT CHECKLIST
════════════════════════════════════════════

Security:
☐ All secrets in Vault (no hardcoded values)
☐ TLS certificates valid (not self-signed in prod)
☐ WAF rules configured
☐ Network policies enforced
☐ RBAC policies configured
☐ Secrets rotation configured
☐ Encryption at rest enabled
☐ API gateway authentication working
☐ Rate limiting active
☐ Brute-force protection enabled

Database:
☐ Replication tested
☐ Backups tested (restore successful)
☐ Indexes created and optimized
☐ Vacuum schedule set
☐ Connection pooling configured
☐ Query logs enabled
☐ Audit logging enabled
☐ Partitioning strategy implemented (if needed)
☐ WAL archiving configured

Messaging:
☐ Kafka topics created with correct replication
☐ Consumer groups configured
☐ Dead-letter queue configured
☐ Message retention policy set
☐ Schema registry configured

Monitoring:
☐ Prometheus scraping all services
☐ Grafana dashboards created
☐ Alert rules configured
☐ On-call schedule active
☐ PagerDuty integration working
☐ Logging pipeline (ELK) working
☐ Tracing (Jaeger) enabled

Performance:
☐ Load testing completed (10x expected traffic)
☐ Latency targets met
☐ Error rate < 0.1%
☐ P99 latency acceptable
☐ Auto-scaling tested
☐ Horizontal pod autoscaling working

Compliance:
☐ Audit logging verified
☐ Data retention policies set
☐ KYC/AML systems working
☐ Sanctions screening enabled
☐ Transaction reporting configured
☐ Privacy controls implemented
☐ GDPR compliance verified

Operations:
☐ Incident response plan finalized
☐ Disaster recovery tested
☐ Runbooks created for common issues
☐ On-call handoff process documented
☐ Deployment automation tested
☐ Rollback procedure tested
☐ Health check endpoints working
☐ Graceful shutdown implemented

Deployment:
☐ Blue-green deployment tested
☐ Canary deployment tested
☐ Rollback procedure tested
☐ Zero-downtime deployment achieved
☐ All services health-check passing
☐ Readiness probes working
☐ Liveness probes working

Final Sign-off:
☐ Security team approval
☐ Compliance team approval
☐ Operations team approval
☐ Product team approval
☐ CFO approval (for financial system)
☐ Legal team approval
```

---

**End of Deployment & Compliance Guide**

## SUMMARY ROADMAP

```
WEEK 1-2: Security Hardening
├─ JWT + Refresh tokens
├─ Brute-force protection
├─ Idempotency keys
└─ Input validation

WEEK 3-4: Fraud & Transaction Safety
├─ Fraud detection rules
├─ Device fingerprinting
├─ Audit logging
└─ Double-entry ledger

MONTH 2: Compliance & Monitoring
├─ KYC/AML setup
├─ Audit logging
├─ Prometheus + Grafana
└─ Centralized logging

MONTH 3: Infrastructure & Scaling
├─ Kubernetes setup
├─ Disaster recovery
├─ Auto-scaling policies
└─ Performance optimization

MONTH 4+: Advanced Features
├─ ML-based fraud detection
├─ Sharding (if 1M+ users)
├─ Multi-region deployment
└─ Advanced analytics

PRODUCTION LAUNCH: Month 4-6
```

