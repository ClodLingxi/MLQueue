"""
MLQueue V2 API 基础使用示例
Python驱动模式：客户端控制训练执行，云端管理配置
"""
from mlqueue import MLQueueV2Client
import time
import random


def train_model(parameters):
    print(f"\n{'=' * 60}")
    print(f"开始训练...")
    print(f"参数: {parameters}")
    print(f"{'=' * 60}")

    # 模拟训练过程
    learning_rate = parameters.get('learning_rate', 0.001)
    epochs = parameters.get('epochs', 10)

    for epoch in range(epochs):
        print(f"Epoch {epoch + 1}/{epochs} - Training...")
        time.sleep(0.3)  # 模拟训练耗时

    # 模拟训练结果
    final_loss = random.uniform(0.1, 0.5) * (1 / learning_rate)
    final_accuracy = random.uniform(0.85, 0.98)

    result = {
        "final_loss": final_loss,
        "accuracy": final_accuracy,
        "parameters": parameters
    }

    print(f"训练完成!")
    print(f"  Loss: {final_loss:.4f}")
    print(f"  Accuracy: {final_accuracy:.4f}")
    print(f"{'=' * 60}\n")

    return result


def main():
    print("=" * 70)
    print("MLQueue V2 API 基础使用示例")
    print("Python驱动模式：客户端控制训练执行")
    print("=" * 70)

    # 1. 连接到云端
    print("\n[1] 连接到MLQueue V2 API...")
    client = MLQueueV2Client(
        api_url="http://localhost:8080/v2",
        api_key="1145"
    )
    print("✓ 连接成功\n")

    # 2. 创建组（代表一个ML项目）
    print("[2] 创建组（ML项目）...")
    group = client.create_group(
        name="MNIST分类项目",
        description="使用不同超参数训练MNIST分类模型"
    )
    print(f"✓ 组创建成功: {group.name} (ID: {group.id})\n")

    # 3. 创建训练单元
    print("[3] 创建训练单元...")
    unit = group.create_training_unit(
        name="CNN超参数搜索",
        config={
            "model": "CNN",
            "dataset": "MNIST",
            "batch_size": 32
        },
        description="搜索最佳学习率"
    )
    print(f"✓ 训练单元创建成功: {unit.name} (ID: {unit.id})")
    print(f"  版本: {unit.version}\n")

    # 4. 添加训练队列（本地添加）
    print("[4] 添加训练队列...")
    queue_configs = [
        {"learning_rate": 0.001, "epochs": 100},
        {"learning_rate": 0.01, "epochs": 20},
        {"learning_rate": 0.1, "epochs": 20},
        {"learning_rate": 0.2, "epochs": 20},
    ]

    for i, params in enumerate(queue_configs, 1):
        queue = unit.add_queue(
            name=f"lr_{params['learning_rate']}",
            parameters=params,
            created_by="client"
        )
        print(f"  ✓ 队列 {i}: {queue.name} (ID: {queue.id})")

    print(f"✓ 共添加 {len(queue_configs)} 个训练队列\n")

    # 5. 启动心跳
    print("[5] 启动心跳保持连接...")
    unit.start_heartbeat(interval=6)  # 每6秒发送一次心跳
    print()

    # 6. 模拟前端添加队列
    print("[6] 模拟前端添加新队列...")
    print("  (在实际使用中，这可以通过Web界面完成)")
    web_queue = unit.add_queue(
        name="lr_0.0001_from_web",
        parameters={"learning_rate": 0.0001, "epochs": 5},
        created_by="web"
    )
    print(f"  ✓ 前端添加的队列: {web_queue.name}\n")

    # 7. 执行训练循环（持续同步并执行新队列）
    print("[7] 开始训练循环（持续同步云端配置）...")
    print("  提示: 按 Ctrl+C 退出循环\n")
    print("=" * 70)

    iteration = 0
    try:
        while True:
            iteration += 1
            print(f"\n>>> 同步循环 #{iteration}")

            # 从云端同步最新配置
            sync_result = unit.sync()

            if sync_result.get("need_sync"):
                print(f"  ✓ 发现云端更新，已同步到版本 v{unit.version}")
            else:
                print(f"  - 无需同步，当前版本: v{unit.version}")

            # 获取所有待执行的队列
            pending_queues = unit.get_pending_queues()

            if not pending_queues:
                print("  - 没有待执行的队列")
                print("  - 等待10秒后继续同步...")
                time.sleep(10)
                continue

            print(f"  - 发现 {len(pending_queues)} 个待执行队列")
            print()

            # 执行待执行队列的首个
            queue = pending_queues[0]
            print(f"  [{i}/{len(pending_queues)}] 开始执行队列: {queue.name}")
            queue.start()

            try:
                result = train_model(queue.parameters)
                queue.complete(
                    result=result,
                    metrics={
                        "accuracy": result["accuracy"],
                        "loss": result["final_loss"]
                    }
                )
                print(f"  ✓ 队列 {queue.name} 执行成功")
            except Exception as e:
                # 标记失败
                error_msg = f"训练失败: {str(e)}"
                queue.fail(error_msg)
                print(f"  ✗ 队列 {queue.name} 执行失败: {error_msg}")

            print("-" * 70)

            print(f"\n本轮训练完成，等待5秒后继续同步...")
            time.sleep(5)

    finally:
        # 停止心跳
        print("\n\n[清理] 停止心跳...")
        unit.stop_heartbeat()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n操作已取消")
    except Exception as e:
        print(f"\n\n错误: {e}")
        import traceback

        traceback.print_exc()
