package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
	"time"
)

func Test_GetState(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()

	assert.NoError(t, err)
	assert.NotNil(t, w)

	assert.Equal(t, StateReady, w.State().Value())
	w.Stop()
	assert.Equal(t, StateStopped, w.State().Value())
}

func Test_Echo(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer w.Stop()

	res, err := w.Exec(&Payload{Body: []byte("hello")})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Head)

	assert.Equal(t, "hello", res.String())
}

func Test_Echo_Slow(t *testing.T) {
	cmd := exec.Command("php", "tests/slow-client.php", "echo", "pipes", "10", "10")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer w.Stop()

	res, err := w.Exec(&Payload{Body: []byte("hello")})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Head)

	assert.Equal(t, "hello", res.String())
}

func Test_Broken(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "broken", "pipes")

	w, err := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		err := w.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "undefined_function()")
	}()
	defer w.Stop()

	res, err := w.Exec(&Payload{Body: []byte("hello")})
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func Test_OnStarted(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "broken", "pipes")
	assert.Nil(t, cmd.Start())

	w, err := newWorker(cmd)
	assert.Nil(t, w)
	assert.NotNil(t, err)

	assert.Equal(t, "can't attach to running process", err.Error())
}

func Test_Error(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "error", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer w.Stop()

	res, err := w.Exec(&Payload{Body: []byte("hello")})
	assert.Nil(t, res)
	assert.NotNil(t, err)

	assert.IsType(t, JobError{}, err)
	assert.Equal(t, "hello", err.Error())
}

func Test_NumExecs(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer w.Stop()

	w.Exec(&Payload{Body: []byte("hello")})
	assert.Equal(t, uint64(1), w.State().NumExecs())

	w.Exec(&Payload{Body: []byte("hello")})
	assert.Equal(t, uint64(2), w.State().NumExecs())

	w.Exec(&Payload{Body: []byte("hello")})
	assert.Equal(t, uint64(3), w.State().NumExecs())
}

func Test_StateUpdated(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer w.Stop()

	tm := time.Now()
	time.Sleep(time.Millisecond)

	w.Exec(&Payload{Body: []byte("hello")})
	assert.True(t, w.State().Updated().After(tm))
}