"""
MLQueue V2 + PyTorch 集成示例
演示如何使用V2 API控制PyTorch训练流程
"""
from mlqueue import MLQueueV2Client, QueueStatus
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, TensorDataset


# 定义简单的神经网络
class SimpleNN(nn.Module):
    def __init__(self, input_size, hidden_size, output_size):
        super(SimpleNN, self).__init__()
        self.fc1 = nn.Linear(input_size, hidden_size)
        self.relu = nn.ReLU()
        self.fc2 = nn.Linear(hidden_size, output_size)

    def forward(self, x):
        x = self.fc1(x)
        x = self.relu(x)
        x = self.fc2(x)
        return x


def create_dummy_dataset(n_samples=1000, input_size=20, output_size=2):
    """创建虚拟数据集用于演示"""
    X = torch.randn(n_samples, input_size)
    y = torch.randint(0, output_size, (n_samples,))
    return TensorDataset(X, y)


def train_with_parameters(parameters):
    """
    使用给定参数训练PyTorch模型

    Args:
        parameters: 训练参数字典，包含:
            - hidden_size: 隐藏层大小
            - learning_rate: 学习率
            - batch_size: 批次大小
            - epochs: 训练轮数

    Returns:
        训练结果字典
    """
    # 提取参数
    hidden_size = parameters.get("hidden_size", 64)
    learning_rate = parameters.get("learning_rate", 0.001)
    batch_size = parameters.get("batch_size", 32)
    epochs = parameters.get("epochs", 5)
    input_size = parameters.get("input_size", 20)
    output_size = parameters.get("output_size", 2)

    print(f"\n{'='*70}")
    print(f"训练配置:")
    print(f"  Hidden Size: {hidden_size}")
    print(f"  Learning Rate: {learning_rate}")
    print(f"  Batch Size: {batch_size}")
    print(f"  Epochs: {epochs}")
    print(f"{'='*70}\n")

    # 创建数据集
    dataset = create_dummy_dataset(n_samples=1000, input_size=input_size, output_size=output_size)
    train_loader = DataLoader(dataset, batch_size=batch_size, shuffle=True)

    # 创建模型
    model = SimpleNN(input_size=input_size, hidden_size=hidden_size, output_size=output_size)
    criterion = nn.CrossEntropyLoss()
    optimizer = optim.Adam(model.parameters(), lr=learning_rate)

    # 训练循环
    best_loss = float('inf')
    epoch_losses = []
    epoch_accuracies = []

    for epoch in range(epochs):
        model.train()
        epoch_loss = 0.0
        correct = 0
        total = 0

        for batch_idx, (data, target) in enumerate(train_loader):
            optimizer.zero_grad()
            output = model(data)
            loss = criterion(output, target)
            loss.backward()
            optimizer.step()

            epoch_loss += loss.item()

            # 计算准确率
            _, predicted = torch.max(output.data, 1)
            total += target.size(0)
            correct += (predicted == target).sum().item()

        # 计算平均损失和准确率
        avg_loss = epoch_loss / len(train_loader)
        accuracy = 100 * correct / total

        epoch_losses.append(avg_loss)
        epoch_accuracies.append(accuracy)

        if avg_loss < best_loss:
            best_loss = avg_loss

        print(f"Epoch [{epoch + 1}/{epochs}] - Loss: {avg_loss:.4f}, Accuracy: {accuracy:.2f}%")

    # 准备返回结果
    result = {
        "best_loss": best_loss,
        "final_loss": epoch_losses[-1],
        "final_accuracy": epoch_accuracies[-1],
        "epoch_losses": epoch_losses,
        "epoch_accuracies": epoch_accuracies
    }

    print(f"\n训练完成!")
    print(f"  最佳Loss: {best_loss:.4f}")
    print(f"  最终准确率: {epoch_accuracies[-1]:.2f}%")
    print(f"{'='*70}\n")

    return result


def main():
    print("="*70)
    print("MLQueue V2 + PyTorch 集成示例")
    print("="*70)

    # 1. 连接到MLQueue V2
    print("\n[1] 连接到MLQueue V2...")
    client = MLQueueV2Client(
        api_url="http://localhost:8080/v2",
        api_key="demo-api-key-12345"
    )
    print("✓ 连接成功\n")

    # 2. 创建项目组
    print("[2] 创建项目组...")
    group = client.create_group(
        name="PyTorch超参数搜索",
        description="神经网络超参数优化实验"
    )
    print(f"✓ 组创建成功: {group.name}\n")

    # 3. 创建训练单元
    print("[3] 创建训练单元...")
    unit = group.create_training_unit(
        name="隐藏层大小 vs 学习率",
        config={
            "model": "SimpleNN",
            "dataset": "synthetic",
            "optimizer": "Adam"
        },
        description="测试不同隐藏层大小和学习率的组合"
    )
    print(f"✓ 训练单元创建成功: {unit.name}\n")

    # 4. 批量添加训练队列
    print("[4] 批量添加训练队列...")

    # 定义超参数网格
    hidden_sizes = [32, 64, 128]
    learning_rates = [0.001, 0.01]

    queue_configs = []
    for hs in hidden_sizes:
        for lr in learning_rates:
            queue_configs.append({
                "name": f"hs{hs}_lr{lr}",
                "parameters": {
                    "hidden_size": hs,
                    "learning_rate": lr,
                    "batch_size": 32,
                    "epochs": 3,
                    "input_size": 20,
                    "output_size": 2
                },
                "created_by": "client"
            })

    queues = unit.add_queues_batch(queue_configs)
    print(f"✓ 批量创建了 {len(queues)} 个训练队列")
    for q in queues:
        print(f"  - {q.name}")
    print()

    # 5. 主动同步
    print("[5] 从云端同步...")
    sync_result = unit.sync()
    print(f"  版本: {unit.version}")
    print(f"  待执行队列数: {len(unit.get_pending_queues())}\n")

    # 6. 执行训练循环
    print("[6] 开始执行训练循环...")
    print("-"*70)

    pending_queues = unit.get_pending_queues()
    total = len(pending_queues)

    for i, queue in enumerate(pending_queues, 1):
        print(f"\n>>> 队列 [{i}/{total}]: {queue.name}")

        # 开始执行
        queue.start()

        try:
            # 执行PyTorch训练
            result = train_with_parameters(queue.parameters)

            # 标记完成
            queue.complete(
                result=result,
                metrics={
                    "best_loss": result["best_loss"],
                    "final_accuracy": result["final_accuracy"]
                }
            )
            print(f"✓ 队列完成: {queue.name}")

        except Exception as e:
            # 标记失败
            import traceback
            error_msg = f"{str(e)}\n{traceback.format_exc()}"
            queue.fail(error_msg)
            print(f"✗ 队列失败: {queue.name}")
            print(f"错误: {e}")

        print("-"*70)

    # 7. 分析结果
    print("\n[7] 分析训练结果...")
    completed_queues = unit.list_queues(status=QueueStatus.COMPLETED)

    print(f"\n已完成队列数: {len(completed_queues)}\n")
    print(f"{'队列名称':<20} {'最佳Loss':<12} {'最终准确率':<12}")
    print("-"*70)

    best_result = None
    best_accuracy = 0

    for q in completed_queues:
        if q.metrics:
            accuracy = q.metrics.get('final_accuracy', 0)
            loss = q.metrics.get('best_loss', 0)
            print(f"{q.name:<20} {loss:<12.4f} {accuracy:<12.2f}%")

            if accuracy > best_accuracy:
                best_accuracy = accuracy
                best_result = {
                    "name": q.name,
                    "parameters": q.parameters,
                    "accuracy": accuracy,
                    "loss": loss
                }

    # 8. 显示最佳结果
    if best_result:
        print(f"\n{'='*70}")
        print("🏆 最佳结果:")
        print(f"  队列: {best_result['name']}")
        print(f"  准确率: {best_result['accuracy']:.2f}%")
        print(f"  损失: {best_result['loss']:.4f}")
        print(f"  参数:")
        for key, value in best_result['parameters'].items():
            print(f"    {key}: {value}")
        print(f"{'='*70}")

    print("\n✓ 所有训练完成!")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n操作已取消")
    except Exception as e:
        print(f"\n\n错误: {e}")
        import traceback
        traceback.print_exc()
