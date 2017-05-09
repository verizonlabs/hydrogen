package test

import (
	mockKv "mesos-framework-sdk/persistence/drivers/etcd/test"
	mockRetry "sprint/task/retry/test"
)

type MockStorage struct {
	mockKv.MockKVStore
	mockRetry.MockRetry
}

type MockBrokenStorage struct {
	mockKv.MockKVStore
	mockRetry.MockBrokenRetry
}
