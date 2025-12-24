#!/usr/bin/env python3
"""
数据库性能测试脚本
用于测试PostgreSQL数据库的查询性能和索引效果

测试内容：
1. 查询优化测试
2. 索引效果验证
3. 复杂查询性能测试
4. 分页查询性能测试
5. 统计查询性能测试

运行方法:
python py_script/db_performance_test.py

需要安装依赖:
pip install psycopg2-binary matplotlib pandas
"""

import psycopg2
import psycopg2.extras
import time
import statistics
import json
import matplotlib.pyplot as plt

from datetime import datetime, timedelta
from typing import List, Dict, Any

# ===== 数据库配置 =====
DB_CONFIG = {
    "host": "localhost",
    "port": 5432,
    "database": "line_management",
    "user": "lineuser",
    "password": "123456"
}

# ===== 测试配置 =====
TEST_ITERATIONS = 10  # 每个查询的测试次数
ENABLE_PLOT = True    # 是否生成图表

class DatabasePerformanceTest:
    """数据库性能测试类"""

    def __init__(self):
        self.conn = None
        self.results = {
            "basic_queries": {},
            "index_tests": {},
            "complex_queries": {},
            "pagination_tests": {},
            "stats_queries": {}
        }

    def connect(self):
        """连接数据库"""
        try:
            self.conn = psycopg2.connect(**DB_CONFIG)
            print("✅ 数据库连接成功")
            return True
        except Exception as e:
            print(f"❌ 数据库连接失败: {e}")
            return False

    def disconnect(self):
        """断开连接"""
        if self.conn:
            self.conn.close()
            print("🔌 数据库连接已断开")

    def execute_query(self, query: str, params: tuple = None, description: str = "") -> Dict[str, Any]:
        """执行查询并测量性能"""
        if not self.conn:
            return {"error": "No database connection"}

        times = []
        results = []

        for i in range(TEST_ITERATIONS):
            try:
                start_time = time.time()
                with self.conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cursor:
                    cursor.execute(query, params)
                    if cursor.description:  # SELECT查询
                        result = cursor.fetchall()
                        results.append(len(result) if result else 0)
                    else:  # 非SELECT查询
                        self.conn.commit()
                        results.append(cursor.rowcount)
                end_time = time.time()
                times.append((end_time - start_time) * 1000)  # 转换为毫秒
            except Exception as e:
                print(f"❌ 查询执行失败: {e}")
                return {"error": str(e)}

        return {
            "avg_time": statistics.mean(times),
            "min_time": min(times),
            "max_time": max(times),
            "median_time": statistics.median(times),
            "std_dev": statistics.stdev(times) if len(times) > 1 else 0,
            "times": times,
            "result_count": results[0] if results else 0,
            "description": description
        }

    def test_basic_queries(self):
        """测试基础查询性能"""
        print(f"\n{'='*60}")
        print("基础查询性能测试")
        print(f"{'='*60}")

        queries = [
            # 用户表查询
            ("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL", None, "用户总数查询"),
            ("SELECT * FROM users WHERE role = 'admin'", None, "管理员查询"),
            ("SELECT * FROM users WHERE is_active = true", None, "活跃用户查询"),

            # 分组表查询
            ("SELECT COUNT(*) FROM groups WHERE deleted_at IS NULL", None, "分组总数查询"),
            ("SELECT * FROM groups WHERE is_active = true", None, "活跃分组查询"),
            ("SELECT * FROM groups WHERE category = 'default'", None, "默认分类分组查询"),

            # Line账号查询
            ("SELECT COUNT(*) FROM line_accounts WHERE deleted_at IS NULL", None, "Line账号总数查询"),
            ("SELECT * FROM line_accounts WHERE online_status = 'online'", None, "在线账号查询"),
            ("SELECT * FROM line_accounts WHERE platform_type = 'line'", None, "Line平台账号查询"),
        ]

        results = {}
        for query, params, description in queries:
            print(f"测试: {description}")
            result = self.execute_query(query, params, description)
            if "error" not in result:
                results[description] = result
                print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
            else:
                print(f"❌ 失败: {result['error']}")

        self.results["basic_queries"] = results

    def test_index_effectiveness(self):
        """测试索引效果"""
        print(f"\n{'='*60}")
        print("索引效果验证测试")
        print(f"{'='*60}")

        # 测试有索引和无索引的查询对比
        index_tests = [
            # 测试activation_code索引
            {
                "name": "activation_code索引测试",
                "indexed": "SELECT * FROM groups WHERE activation_code = 'TEST001'",
                "non_indexed": "SELECT * FROM groups WHERE activation_code LIKE '%TEST001%'"
            },

            # 测试group_id索引
            {
                "name": "group_id索引测试",
                "indexed": "SELECT COUNT(*) FROM line_accounts WHERE group_id = 1",
                "non_indexed": "SELECT COUNT(*) FROM line_accounts WHERE group_id::text = '1'"
            },

            # 测试时间范围索引
            {
                "name": "时间索引测试",
                "indexed": "SELECT COUNT(*) FROM incoming_logs WHERE incoming_time >= '2025-01-01' AND incoming_time < '2025-02-01'",
                "non_indexed": "SELECT COUNT(*) FROM incoming_logs WHERE EXTRACT(YEAR FROM incoming_time) = 2025 AND EXTRACT(MONTH FROM incoming_time) = 1"
            }
        ]

        results = {}
        for test in index_tests:
            print(f"测试: {test['name']}")

            # 测试有索引的查询
            print("  - 有索引查询:")
            indexed_result = self.execute_query(test['indexed'], description=f"{test['name']}-indexed")
            if "error" not in indexed_result:
                print(f"  平均时间: {indexed_result.get('avg_time', 0):.2f}ms")
            else:
                print(f"    ❌ 失败: {indexed_result['error']}")

            # 测试无索引的查询
            print("  - 无索引查询:")
            non_indexed_result = self.execute_query(test['non_indexed'], description=f"{test['name']}-non-indexed")
            if "error" not in non_indexed_result:
                print(f"  平均时间: {non_indexed_result.get('avg_time', 0):.2f}ms")
                if indexed_result.get('avg_time', 0) > 0:
                    speedup = non_indexed_result['avg_time'] / indexed_result['avg_time']
                    print(f"    性能提升: {speedup:.1f}x")
            else:
                print(f"    ❌ 失败: {non_indexed_result['error']}")

            results[test['name']] = {
                "indexed": indexed_result,
                "non_indexed": non_indexed_result
            }

        self.results["index_tests"] = results

    def test_complex_queries(self):
        """测试复杂查询性能"""
        print(f"\n{'='*60}")
        print("复杂查询性能测试")
        print(f"{'='*60}")

        queries = [
            # JOIN查询
            ("""
                SELECT g.activation_code, g.remark, COUNT(la.id) as account_count,
                       COALESCE(gs.total_incoming, 0) as total_incoming
                FROM groups g
                LEFT JOIN line_accounts la ON la.group_id = g.id AND la.deleted_at IS NULL
                LEFT JOIN group_stats gs ON gs.group_id = g.id
                WHERE g.deleted_at IS NULL AND g.is_active = true
                GROUP BY g.id, g.activation_code, g.remark, gs.total_incoming
                ORDER BY account_count DESC
            """, None, "分组账号统计JOIN查询"),

            # 子查询
            ("""
                SELECT * FROM line_accounts
                WHERE group_id IN (
                    SELECT id FROM groups
                    WHERE is_active = true AND deleted_at IS NULL
                )
                AND deleted_at IS NULL
            """, None, "子查询-活跃分组的账号"),

            # 窗口函数
            ("""
                SELECT activation_code, remark,
                       ROW_NUMBER() OVER (ORDER BY created_at) as row_num,
                       RANK() OVER (ORDER BY created_at) as rank_num
                FROM groups
                WHERE deleted_at IS NULL
                ORDER BY created_at
            """, None, "窗口函数-分组排名"),

            # JSON查询
            ("""
                SELECT * FROM customers
                WHERE tags::text != 'null'
                AND deleted_at IS NULL
                LIMIT 100
            """, None, "JSON字段查询"),
        ]

        results = {}
        for query, params, description in queries:
            print(f"测试: {description}")
            result = self.execute_query(query, params, description)
            if "error" not in result:
                results[description] = result
                print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
            else:
                print(f"❌ 失败: {result['error']}")

        self.results["complex_queries"] = results

    def test_pagination_queries(self):
        """测试分页查询性能"""
        print(f"\n{'='*60}")
        print("分页查询性能测试")
        print(f"{'='*60}")

        # 测试不同分页大小的性能
        page_sizes = [10, 50, 100, 500, 1000]
        results = {}

        for page_size in page_sizes:
            print(f"测试分页大小: {page_size}")

            # 分页查询line_accounts表
            query = f"""
                SELECT * FROM line_accounts
                WHERE deleted_at IS NULL
                ORDER BY created_at DESC
                LIMIT {page_size} OFFSET 0
            """

            result = self.execute_query(query, description=f"分页查询-{page_size}")
            if "error" not in result:
                results[f"page_size_{page_size}"] = result
                print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
            else:
                print(f"❌ 失败: {result['error']}")

        self.results["pagination_tests"] = results

    def test_stats_queries(self):
        """测试统计查询性能"""
        print(f"\n{'='*60}")
        print("统计查询性能测试")
        print(f"{'='*60}")

        queries = [
            # 基础统计
            ("SELECT COUNT(*) FROM groups WHERE deleted_at IS NULL", None, "分组总数统计"),
            ("SELECT COUNT(*) FROM line_accounts WHERE deleted_at IS NULL", None, "账号总数统计"),
            ("SELECT COUNT(*) FROM customers WHERE deleted_at IS NULL", None, "客户总数统计"),

            # 聚合统计
            ("""
                SELECT platform_type, COUNT(*) as count
                FROM line_accounts
                WHERE deleted_at IS NULL
                GROUP BY platform_type
            """, None, "按平台统计账号"),

            ("""
                SELECT online_status, COUNT(*) as count
                FROM line_accounts
                WHERE deleted_at IS NULL
                GROUP BY online_status
            """, None, "按状态统计账号"),

            # 时间范围统计
            ("""
                SELECT DATE(incoming_time), COUNT(*) as daily_count
                FROM incoming_logs
                WHERE incoming_time >= CURRENT_DATE - INTERVAL '30 days'
                GROUP BY DATE(incoming_time)
                ORDER BY DATE(incoming_time)
            """, None, "最近30天每日进线统计"),

            # 复杂统计查询
            ("""
                SELECT
                    g.activation_code,
                    COUNT(DISTINCT la.id) as accounts,
                    COUNT(DISTINCT c.id) as customers,
                    COUNT(il.id) as incoming_count
                FROM groups g
                LEFT JOIN line_accounts la ON la.group_id = g.id AND la.deleted_at IS NULL
                LEFT JOIN customers c ON c.group_id = g.id AND c.deleted_at IS NULL
                LEFT JOIN incoming_logs il ON il.group_id = g.id AND il.incoming_time >= CURRENT_DATE - INTERVAL '7 days'
                WHERE g.deleted_at IS NULL
                GROUP BY g.id, g.activation_code
                ORDER BY accounts DESC
            """, None, "分组综合统计（7天）"),
        ]

        results = {}
        for query, params, description in queries:
            print(f"测试: {description}")
            result = self.execute_query(query, params, description)
            if "error" not in result:
                results[description] = result
                print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
            else:
                print(f"❌ 失败: {result['error']}")

        self.results["stats_queries"] = results

    def analyze_query_performance(self):
        """分析查询性能"""
        print(f"\n{'='*80}")
        print("📊 数据库性能分析报告")
        print(f"{'='*80}")

        # 基础查询分析
        basic = self.results.get("basic_queries", {})
        if basic:
            print("\n🔍 基础查询性能:")
            slow_queries = []
            for name, result in basic.items():
                avg_time = result.get('avg_time', 0)
                print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
                if avg_time > 100:  # 超过100ms认为是慢查询
                    slow_queries.append((name, avg_time))

            if slow_queries:
                print("\n🐌 慢查询警告:")
                for name, time_taken in slow_queries:
                    print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
        # 索引效果分析
        index_tests = self.results.get("index_tests", {})
        if index_tests:
            print("\n📈 索引效果分析:")
            for test_name, test_results in index_tests.items():
                indexed = test_results.get("indexed", {})
                non_indexed = test_results.get("non_indexed", {})

                if "error" not in indexed and "error" not in non_indexed:
                    indexed_time = indexed.get('avg_time', 0)
                    non_indexed_time = non_indexed.get('avg_time', 0)

                    if indexed_time > 0 and non_indexed_time > 0:
                        speedup = non_indexed_time / indexed_time
                        print(f"    性能提升: {speedup:.1f}x")
                        if speedup > 5:
                            print("  ✅ 索引效果显著")
                        elif speedup > 2:
                            print("  ⚠️  索引效果一般")
                        else:
                            print("  ❌ 索引效果不明显")
        # 复杂查询分析
        complex_queries = self.results.get("complex_queries", {})
        if complex_queries:
            print("\n🔄 复杂查询性能:")
            for name, result in complex_queries.items():
                avg_time = result.get('avg_time', 0)
                print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
        # 分页查询分析
        pagination = self.results.get("pagination_tests", {})
        if pagination:
            print("\n📄 分页查询性能:")
            for page_size, result in pagination.items():
                avg_time = result.get('avg_time', 0)
                size = page_size.replace("page_size_", "")
                print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
        # 统计查询分析
        stats = self.results.get("stats_queries", {})
        if stats:
            print("\n📊 统计查询性能:")
            for name, result in stats.items():
                avg_time = result.get('avg_time', 0)
                print(f"  平均时间: {result.get('avg_time', 0):.2f}ms")
    def generate_performance_report(self):
        """生成性能报告"""
        print("\n📝 生成性能测试报告...")
        report = {
            "test_time": datetime.now().isoformat(),
            "database_config": {k: v for k, v in DB_CONFIG.items() if k != "password"},
            "test_config": {
                "iterations": TEST_ITERATIONS,
                "enable_plot": ENABLE_PLOT
            },
            "results": self.results
        }

        # 保存JSON报告
        with open("db_performance_report.json", "w", encoding="utf-8") as f:
            json.dump(report, f, indent=2, ensure_ascii=False)

        # 生成性能对比图表
        if ENABLE_PLOT:
            self.generate_performance_charts()

        print("✅ 性能测试报告已生成: db_performance_report.json")
        if ENABLE_PLOT:
            print("📊 性能图表已生成: db_performance_charts.png")

    def generate_performance_charts(self):
        """生成性能对比图表"""
        try:
            # 收集所有查询的性能数据
            query_names = []
            avg_times = []

            # 基础查询
            for name, result in self.results.get("basic_queries", {}).items():
                if "error" not in result:
                    query_names.append(f"基础-{name[:20]}")
                    avg_times.append(result.get('avg_time', 0))

            # 复杂查询
            for name, result in self.results.get("complex_queries", {}).items():
                if "error" not in result:
                    query_names.append(f"复杂-{name[:20]}")
                    avg_times.append(result.get('avg_time', 0))

            # 统计查询
            for name, result in self.results.get("stats_queries", {}).items():
                if "error" not in result:
                    query_names.append(f"统计-{name[:20]}")
                    avg_times.append(result.get('avg_time', 0))

            if query_names and avg_times:
                # 创建图表
                plt.figure(figsize=(15, 8))

                # 主要图表
                plt.subplot(2, 1, 1)
                bars = plt.bar(range(len(query_names)), avg_times, color='skyblue', alpha=0.8)
                plt.xlabel('查询类型')
                plt.ylabel('平均响应时间 (ms)')
                plt.title('数据库查询性能对比')
                plt.xticks(range(len(query_names)), query_names, rotation=45, ha='right')
                plt.grid(True, alpha=0.3)

                # 添加数值标签
                for bar, time_val in zip(bars, avg_times):
                    plt.text(bar.get_x() + bar.get_width()/2, bar.get_y() + bar.get_height() + max(avg_times)*0.01,
                           '.1f', ha='center', va='bottom', fontsize=8)

                # 分页查询子图
                plt.subplot(2, 1, 2)
                pagination = self.results.get("pagination_tests", {})
                if pagination:
                    page_sizes = []
                    page_times = []
                    for page_size, result in pagination.items():
                        if "error" not in result:
                            size = int(page_size.replace("page_size_", ""))
                            page_sizes.append(size)
                            page_times.append(result.get('avg_time', 0))

                    if page_sizes and page_times:
                        plt.plot(page_sizes, page_times, 'ro-', linewidth=2, markersize=8)
                        plt.xlabel('分页大小')
                        plt.ylabel('平均响应时间 (ms)')
                        plt.title('分页查询性能趋势')
                        plt.grid(True, alpha=0.3)
                        plt.xticks(page_sizes)

                        # 添加数值标签
                        for x, y in zip(page_sizes, page_times):
                            plt.text(x, y + max(page_times)*0.02, '.1f', ha='center', va='bottom')

                plt.tight_layout()
                plt.savefig("db_performance_charts.png", dpi=150, bbox_inches='tight')
                plt.close()

        except Exception as e:
            print(f"❌ 生成图表失败: {e}")

    def run_full_test(self):
        """运行完整的数据库性能测试"""
        print("🚀 数据库性能测试开始")
        print("=" * 80)
        print(f"数据库: {DB_CONFIG['database']}")
        print(f"主机: {DB_CONFIG['host']}:{DB_CONFIG['port']}")
        print(f"测试次数: {TEST_ITERATIONS}")
        print("=" * 80)

        if not self.connect():
            return

        try:
            # 1. 基础查询测试
            self.test_basic_queries()

            # 2. 索引效果验证
            self.test_index_effectiveness()

            # 3. 复杂查询测试
            self.test_complex_queries()

            # 4. 分页查询测试
            self.test_pagination_queries()

            # 5. 统计查询测试
            self.test_stats_queries()

            # 6. 性能分析
            self.analyze_query_performance()

            # 7. 生成报告
            self.generate_performance_report()

        except Exception as e:
            print(f"❌ 测试过程中出错: {e}")
        finally:
            self.disconnect()

        print("✅ 数据库性能测试完成！")
        print("详细报告: db_performance_report.json")
        print("性能图表: db_performance_charts.png")

def main():
    """主函数"""
    print("数据库性能测试工具")
    print("此工具将测试PostgreSQL数据库的查询性能和索引效果")
    print("-" * 60)
    print("测试内容:")
    print("1. 基础查询性能测试")
    print("2. 索引效果验证")
    print("3. 复杂查询性能测试")
    print("4. 分页查询性能测试")
    print("5. 统计查询性能测试")
    print("-" * 60)

    # 确认开始测试
    try:
        response = input("是否开始数据库性能测试？(y/N): ").strip().lower()
        if response != 'y':
            print("测试已取消")
            return
    except KeyboardInterrupt:
        print("\n测试已取消")
        return

    # 创建并运行性能测试
    tester = DatabasePerformanceTest()
    tester.run_full_test()

if __name__ == "__main__":
    main()
