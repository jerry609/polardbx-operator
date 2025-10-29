/*
Copyright 2021 Alibaba Group Holding Limited.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package control

import (
	"time"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
)

// NewStandardRateLimiter creates a standard rate limiter with 60 qps and 10 bucket size.
// This is used by most controllers with a max exponential backoff of 300 seconds.
func NewStandardRateLimiter() workqueue.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 300*time.Second),
		// 60 qps, 10 bucket size. This is only for retry speed. It's only the overall factor (not per item).
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(60), 10)},
	)
}

// NewLogCollectorRateLimiter creates a rate limiter optimized for log collector controllers
// with 10 qps and 1 bucket size, and a max exponential backoff of 5 seconds.
func NewLogCollectorRateLimiter() workqueue.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 5*time.Second),
		// 10 qps, 1 bucket size. This is only for retry speed. It's only the overall factor (not per item).
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 1)},
	)
}

// NewXStoreRateLimiter creates a rate limiter for XStore controllers
// with 60 qps and 10 bucket size, and a max exponential backoff of 60 seconds.
func NewXStoreRateLimiter() workqueue.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 60*time.Second),
		// 60 qps, 10 bucket size. This is only for retry speed. It's only the overall factor (not per item).
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(60), 10)},
	)
}

// NewXStoreFollowerRateLimiter creates a rate limiter for XStore follower controllers
// with 60 qps and 10 bucket size, and a max exponential backoff of 30 seconds.
func NewXStoreFollowerRateLimiter() workqueue.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 30*time.Second),
		// 60 qps, 10 bucket size. This is only for retry speed. It's only the overall factor (not per item).
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(60), 10)},
	)
}
