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
	"testing"
	"time"

	"k8s.io/client-go/util/workqueue"
)

func TestNewStandardRateLimiter(t *testing.T) {
	limiter := NewStandardRateLimiter()
	if limiter == nil {
		t.Fatal("NewStandardRateLimiter returned nil")
	}
	
	// Test that it's a valid rate limiter
	_ = limiter.When("test-item")
	limiter.Forget("test-item")
	limiter.NumRequeues("test-item")
}

func TestNewLogCollectorRateLimiter(t *testing.T) {
	limiter := NewLogCollectorRateLimiter()
	if limiter == nil {
		t.Fatal("NewLogCollectorRateLimiter returned nil")
	}
	
	// Test that it's a valid rate limiter
	_ = limiter.When("test-item")
	limiter.Forget("test-item")
	limiter.NumRequeues("test-item")
}

func TestNewXStoreRateLimiter(t *testing.T) {
	limiter := NewXStoreRateLimiter()
	if limiter == nil {
		t.Fatal("NewXStoreRateLimiter returned nil")
	}
	
	// Test that it's a valid rate limiter
	_ = limiter.When("test-item")
	limiter.Forget("test-item")
	limiter.NumRequeues("test-item")
}

func TestNewXStoreFollowerRateLimiter(t *testing.T) {
	limiter := NewXStoreFollowerRateLimiter()
	if limiter == nil {
		t.Fatal("NewXStoreFollowerRateLimiter returned nil")
	}
	
	// Test that it's a valid rate limiter
	_ = limiter.When("test-item")
	limiter.Forget("test-item")
	limiter.NumRequeues("test-item")
}

func TestRateLimiterBackoff(t *testing.T) {
	limiter := NewStandardRateLimiter()
	
	// First failure should have minimal backoff
	delay1 := limiter.When("test-item")
	if delay1 > 100*time.Millisecond {
		t.Errorf("First retry delay too long: %v", delay1)
	}
	
	// Simulate multiple failures to test exponential backoff
	for i := 0; i < 10; i++ {
		_ = limiter.When("test-item")
	}
	
	// After many failures, should still be capped at max backoff (300s)
	delay := limiter.When("test-item")
	if delay > 300*time.Second {
		t.Errorf("Delay exceeded maximum: %v", delay)
	}
	
	// Forget the item
	limiter.Forget("test-item")
	
	// After forgetting, should reset to minimal backoff
	delayReset := limiter.When("test-item")
	if delayReset > 100*time.Millisecond {
		t.Errorf("Reset delay too long: %v", delayReset)
	}
}

func TestLogCollectorRateLimiterLowerMaxBackoff(t *testing.T) {
	limiter := NewLogCollectorRateLimiter()
	
	// Simulate multiple failures to test exponential backoff
	for i := 0; i < 20; i++ {
		_ = limiter.When("test-item")
	}
	
	// After many failures, should be capped at 5s (not 300s)
	delay := limiter.When("test-item")
	if delay > 5*time.Second {
		t.Errorf("Delay exceeded log collector maximum: %v", delay)
	}
}

func TestXStoreRateLimiterMaxBackoff(t *testing.T) {
	limiter := NewXStoreRateLimiter()
	
	// Simulate multiple failures to test exponential backoff
	for i := 0; i < 20; i++ {
		_ = limiter.When("test-item")
	}
	
	// After many failures, should be capped at 60s
	delay := limiter.When("test-item")
	if delay > 60*time.Second {
		t.Errorf("Delay exceeded XStore maximum: %v", delay)
	}
}

func TestRateLimiterIsMaxOfRateLimiter(t *testing.T) {
	// Test that our rate limiters implement the expected type
	limiters := []workqueue.RateLimiter{
		NewStandardRateLimiter(),
		NewLogCollectorRateLimiter(),
		NewXStoreRateLimiter(),
		NewXStoreFollowerRateLimiter(),
	}
	
	for _, limiter := range limiters {
		if limiter == nil {
			t.Fatal("Rate limiter is nil")
		}
	}
}
