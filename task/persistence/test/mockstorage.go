package test

import (
	mockKv "mesos-framework-sdk/persistence/drivers/etcd/test"
	mockRetry "mesos-framework-sdk/task/retry/test"
)

type MockStorage struct {
	mockKv.MockKVStore
	mockRetry.MockRetry
}

type MockBrokenStorage struct {
	mockKv.MockBrokenKVStore
	mockRetry.MockBrokenRetry
}
