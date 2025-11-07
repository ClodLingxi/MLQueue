"""
测试 TrainingConfig 类
"""
import unittest
from mlqueue.config import TrainingConfig


class TestTrainingConfig(unittest.TestCase):
    """TrainingConfig 测试类"""

    def test_init_with_dict(self):
        """测试使用字典初始化"""
        config_dict = {"hidden_size": 64, "learning_rate": 0.001}
        config = TrainingConfig(config_dict)

        self.assertEqual(config.get("hidden_size"), 64)
        self.assertEqual(config.get("learning_rate"), 0.001)

    def test_init_with_kwargs(self):
        """测试使用关键字参数初始化"""
        config = TrainingConfig(hidden_size=128, epochs=100)

        self.assertEqual(config.get("hidden_size"), 128)
        self.assertEqual(config.get("epochs"), 100)

    def test_set_and_get(self):
        """测试设置和获取配置项"""
        config = TrainingConfig()
        config.set("batch_size", 32)

        self.assertEqual(config.get("batch_size"), 32)
        self.assertIsNone(config.get("nonexistent"))
        self.assertEqual(config.get("nonexistent", "default"), "default")

    def test_update(self):
        """测试批量更新"""
        config = TrainingConfig({"a": 1})
        config.update({"b": 2, "c": 3})

        self.assertEqual(config.get("a"), 1)
        self.assertEqual(config.get("b"), 2)
        self.assertEqual(config.get("c"), 3)

    def test_to_dict(self):
        """测试转换为字典"""
        config = TrainingConfig({"hidden_size": 64})
        config.config_id = "test_id"

        config_dict = config.to_dict()

        self.assertEqual(config_dict["config_id"], "test_id")
        self.assertEqual(config_dict["config"]["hidden_size"], 64)
        self.assertIn("created_at", config_dict)

    def test_from_dict(self):
        """测试从字典创建"""
        data = {
            "config_id": "test_123",
            "config": {"hidden_size": 128},
            "created_at": "2024-01-01T00:00:00"
        }

        config = TrainingConfig.from_dict(data)

        self.assertEqual(config.config_id, "test_123")
        self.assertEqual(config.get("hidden_size"), 128)
        self.assertEqual(config.created_at, "2024-01-01T00:00:00")

    def test_chain_operations(self):
        """测试链式调用"""
        config = TrainingConfig().set("a", 1).set("b", 2).update({"c": 3})

        self.assertEqual(config.get("a"), 1)
        self.assertEqual(config.get("b"), 2)
        self.assertEqual(config.get("c"), 3)


if __name__ == '__main__':
    unittest.main()
