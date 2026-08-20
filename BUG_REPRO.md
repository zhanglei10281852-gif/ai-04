# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

同一工作空间先用旧 schema family 登记了一个 source revision，切到新版 schema 后再次提交该 revision，服务返回成功和新的 snapshot ID，可目录里仍只有旧记录，后续拿新 ID 会查不到对象。当前只排查原因，不改代码；生产代码、测试与配置继续保留原状。请追踪唯一约束错误从存储层传到 service 返回值的过程，说明这个未落库却被当成成功的对象是怎么出现的。

## 含 Bug 版本

- 仓库：zhanglei10281852-gif/ai-04
- 仓库地址：https://github.com/zhanglei10281852-gif/ai-04.git
- parent SHA：c37e98cd074cf0ee96571f111fed6a3aec2829c6

## 复现步骤

```bash
git clone -- https://github.com/zhanglei10281852-gif/ai-04.git bug-repro
cd bug-repro
git checkout --detach c37e98cd074cf0ee96571f111fed6a3aec2829c6
go test ./internal/service -run ^TestDuplicateSnapshotRegistrationDoesNotReturnGhostRecord$ -count=1
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/service -run ^TestDuplicateSnapshotRegistrationDoesNotReturnGhostRecord$ -count=1
--- FAIL: TestDuplicateSnapshotRegistrationDoesNotReturnGhostRecord (0.52s)
    annotation_core_behavior_test.go:68: duplicate registration error = <nil>
FAIL
FAIL	github.com/zhanglei10281852-gif/ai/internal/service	0.528s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/service -run ^TestDuplicateSnapshotRegistrationDoesNotReturnGhostRecord$ -count=1
--- FAIL: TestDuplicateSnapshotRegistrationDoesNotReturnGhostRecord (1.11s)
    annotation_core_behavior_test.go:68: duplicate registration error = <nil>
FAIL
FAIL	github.com/zhanglei10281852-gif/ai/internal/service	1.294s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

诊断必须追踪 internal/storage/sqlite/write_core.go 的 InsertDatasetSnapshot 与 internal/service/catalog.go 的 CatalogService.RegisterSnapshot，证明同 revision、不同 schema family 的唯一约束冲突被包装为 domain.ErrAlreadyExists，随后又被服务当作幂等成功并返回未持久化的请求对象；需以事务未新增目录行和返回的新 ID 无法查询为证据。定向复现完成且仓库零改动。
