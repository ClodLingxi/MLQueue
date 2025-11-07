# MLQueue Python SDK

Python client library for MLQueue task management system.

## Installation

```bash
pip install -e .
```

## Quick Start

### V2 API (Recommended)

```python
from mlqueue import MLQueueV2Client

# Initialize client
client = MLQueueV2Client(
    api_url="http://localhost:8080",
    api_key="your-api-key"
)

# Create group and training unit
group = client.create_group(name="My Project", description="Description")
unit = client.create_training_unit(
    group_id=group.group_id,
    name="ResNet50",
    config={"model": "resnet50", "dataset": "imagenet"}
)

# Create queue
queue = client.create_queue(
    unit_id=unit.unit_id,
    parameters={"lr": 0.001, "batch_size": 64, "epochs": 100}
)

# Execute training
client.start_queue(queue.queue_id)
result = {"accuracy": 0.95, "loss": 0.123}
client.complete_queue(queue.queue_id, result=result, metrics=result)
```

### V2 with PyTorch

```python
from mlqueue import MLQueueV2Client
import torch
import torch.nn as nn

client = MLQueueV2Client(api_url="http://localhost:8080", api_key="your-api-key")

# Setup
group = client.create_group(name="PyTorch Training")
unit = client.create_training_unit(
    group_id=group.group_id,
    name="CNN",
    config={"model_type": "cnn"}
)

# Add training tasks
for lr in [0.001, 0.01, 0.1]:
    client.create_queue(
        unit_id=unit.unit_id,
        parameters={"lr": lr, "epochs": 10}
    )

# Execute queues
client.sync_unit(unit.unit_id)
queues = client.list_queues(unit.unit_id, status="pending")

for q in queues:
    client.start_queue(q.queue_id)

    # Your training code
    model = YourModel()
    optimizer = torch.optim.Adam(model.parameters(), lr=q.parameters["lr"])
    # ... training loop ...

    result = {"final_loss": 0.123, "accuracy": 0.95}
    client.complete_queue(q.queue_id, result=result)
```

### V1 API (Legacy)

```python
from mlqueue import MLTrainer

trainer = MLTrainer(
    api_url="http://localhost:8080/v1",
    api_key="your-api-key"
)

# Create task
config = {"lr": 0.001, "batch_size": 32, "epochs": 100}
task = trainer.add_config(
    name="Training Experiment",
    config=config,
    priority=1
)

# Execute with context manager
with trainer.start_training("experiment", config) as ctx:
    cfg = ctx.get_config()
    # Train your model
    result = {"loss": 0.123, "accuracy": 0.95}
    ctx.log_result(result)
```

### V1 Batch Operations

```python
# Batch create tasks
configs = [
    {"name": "lr_0.001", "config": {"lr": 0.001}, "priority": 1},
    {"name": "lr_0.01", "config": {"lr": 0.01}, "priority": 2},
    {"name": "lr_0.1", "config": {"lr": 0.1}, "priority": 3},
]

tasks = trainer.add_configs(configs)

# List tasks
all_tasks = trainer.list_tasks()
completed_tasks = trainer.list_tasks(status="completed")

# Queue management
status = trainer.get_queue_status()
trainer.queue.update_priority(task_id, priority=10)
trainer.queue.cancel_task(task_id)
```

## Environment Variables

```bash
export MLQUEUE_API_URL="http://localhost:8080"
export MLQUEUE_API_KEY="your-api-key"
```

## API Reference

### MLQueueV2Client

```python
client = MLQueueV2Client(api_url, api_key)

# Groups
client.create_group(name, description)
client.list_groups()

# Training Units
client.create_training_unit(group_id, name, config)
client.get_training_unit(unit_id)
client.sync_unit(unit_id)
client.update_heartbeat(unit_id)

# Queues
client.create_queue(unit_id, parameters, order, created_by)
client.list_queues(unit_id, status)
client.start_queue(queue_id)
client.complete_queue(queue_id, result, metrics)
client.fail_queue(queue_id, error_message)
```

### MLTrainer (V1)

```python
trainer = MLTrainer(api_url, api_key, queue_name, auto_upload)

# Task management
trainer.add_config(name, config, priority)
trainer.add_configs(configs)
trainer.list_tasks(status, limit, offset)
trainer.start_training(name, config, priority)

# Queue operations
trainer.get_queue_status()
trainer.queue.update_priority(task_id, priority)
trainer.queue.cancel_task(task_id)
trainer.queue.reorder(task_ids)
```

## Examples

See `examples/` directory:
- `v2_basic_usage.py` - V2 API basics
- `v2_pytorch_example.py` - PyTorch integration

## License

MIT
