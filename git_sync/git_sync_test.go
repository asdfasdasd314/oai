package git_sync

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInsertion(t *testing.T) {
	queue := []int64{}
	insertIntoQueue(0, &queue)
	fmt.Println(queue)
	correctQueue := []int64{0}
	assert.Equal(t, correctQueue, queue, "These should be equal")

	insertIntoQueue(1, &queue)
	correctQueue = []int64{0, 1}
	assert.Equal(t, correctQueue, queue, "These should be equal")

	insertIntoQueue(3, &queue)
	correctQueue = []int64{0, 1, 3}
	assert.Equal(t, correctQueue, queue, "These should be equal")

	insertIntoQueue(2, &queue)
	correctQueue = []int64{0, 1, 2, 3}
	assert.Equal(t, correctQueue, queue, "These should be equal")
}
