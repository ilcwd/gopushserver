package gopushserver

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

const (
	TEST_DATA_PATTERN = "_DaTa_%d"
)

func newDataN(n int) []byte {
	return []byte(fmt.Sprintf(TEST_DATA_PATTERN, n))
}

func isDataN(src []byte, n int) bool {
	dataN := newDataN(n)
	return (bytes.Compare(dataN, src) == 0)
}

func TestMessageQueue(t *testing.T) {
	var (
		QUEUE_SIZE           int           = 7
		TIME_DELTA           time.Duration = time.Millisecond * 100
		MESSAGE_EXPIRES      time.Duration = time.Second
		FISRT_ENQUEUE_AMOUNT int           = 3
		FISRT_SLEEP_TIME     time.Duration = time.Second / 2
		// SECOND_ENQUEUE_AMOUNT  int           = QUEUE_SIZE - FISRT_ENQUEUE_AMOUNT
		THIRD_ENQUEUE_AMOUNT int           = 5
		RECYCLE_TIME         time.Duration = time.Second * 5
	)

	q := NewQueue(QUEUE_SIZE, MESSAGE_EXPIRES, RECYCLE_TIME)
	for i := 0; i < FISRT_ENQUEUE_AMOUNT; i++ {
		q.Enqueue(newDataN(i))
	}

	time.Sleep(FISRT_SLEEP_TIME)

	for i := FISRT_ENQUEUE_AMOUNT; i < QUEUE_SIZE; i++ {
		q.Enqueue(newDataN(i))
	}

	// enqueue when overflow, the first item is removed.
	q.Enqueue(newDataN(QUEUE_SIZE))

	// all data are stored as expected
	for i, m := range q.Messages {
		// the first item is newly enqueued
		if i == 0 {
			i = QUEUE_SIZE
		}
		if !isDataN(m.Body, i) {
			t.Fatalf("unexpected data %s, %d\n", string(m.Body), i)
		}
	}

	// deque some data from the first part.
	for i := 0; i < FISRT_ENQUEUE_AMOUNT-1; i++ {
		data := q.Dequeue()
		if !isDataN(data, i+1) {
			t.Fatalf("unexpected data %s, %d\n", string(data), i)
		}
	}

	// wait for message of first part expired
	time.Sleep(MESSAGE_EXPIRES - TIME_DELTA)

	// now all message in the first part are expired,
	// dequeue some message from second part to verify it.
	for i := FISRT_ENQUEUE_AMOUNT; i < QUEUE_SIZE-1; i++ {
		data := q.Dequeue()
		if !isDataN(data, i) {
			t.Fatalf("unexpected data %s, %d\n", string(data), i)
		}
	}

	// verify current pointer and queue length
	if q.Current != QUEUE_SIZE-1 {
		t.Fatalf("unexpected current pointer %d", q.Current)
	}
	// second part left 1, and the overflow 1
	if q.Length != 2 {
		t.Fatalf("unexpected queue length %d", q.Length)
	}

	// enqueue third round of data
	for i := QUEUE_SIZE; i < QUEUE_SIZE+THIRD_ENQUEUE_AMOUNT; i++ {
		q.Enqueue(newDataN(i + 1))
	}

	// verify current pointer and queue length
	if q.Current != QUEUE_SIZE-1 {
		t.Fatalf("unexpected current pointer %d", q.Current)
	}
	if q.Length != 2+THIRD_ENQUEUE_AMOUNT {
		t.Fatalf("unexpected queue length %d", q.Length)
	}

	// dequeue and leave only 2 items
	for i := q.Current; q.Length > 2; i++ {
		data := q.Dequeue()
		if !isDataN(data, i) {
			t.Fatalf("unexpected data %s, %d\n", string(data), i)
		}
	}

	if q.Length != 2 {
		t.Fatalf("unexpected queue length %d", q.Length)
	}

	// wait for all item expire and recycle
	time.Sleep(RECYCLE_TIME)

	// verify current pointer and queue length
	if q.Current != 0 {
		t.Fatalf("unexpected current pointer %d", q.Current)
	}
	if q.Length != 0 {
		t.Fatalf("unexpected queue length %d", q.Length)
	}

}
