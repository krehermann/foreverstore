package store

import (
	"bytes"
	"fmt"
	"hash"
	"io"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBlobStore_Create(t *testing.T) {
	type fields struct {
		config     BlobStoreConfig
		registerCh chan<- *ObjectRef
	}
	type args struct {
		key string
		r   io.Reader
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *ObjectRef
		wantBytes []byte
		wantErr   bool
	}{
		{
			name: "create 1",
			fields: fields{
				config: BlobStoreConfig{
					PathFunc: ContentPath,
					Root:     t.TempDir(),
					Logger:   zap.Must(zap.NewDevelopment()),
				},
				registerCh: make(chan<- *ObjectRef),
			},
			args: args{
				key: "key",
				r:   bytes.NewReader([]byte("some content")),
			},
			want: &ObjectRef{
				Key: "key",
			},
			wantBytes: []byte("some content"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//create the store
			str, err := NewBlobStore(tt.fields.config)
			require.NoError(t, err, tt.name)
			// create a blob
			b, err := str.Create(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileStore.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// write the blob
			d, err := io.ReadAll(tt.args.r)
			assert.NoError(t, err)
			_, err = b.Write(d)
			assert.NoError(t, err)
			assert.NoError(t, b.Close())

			// stat it to get name
			_, err = b.Stat()
			assert.NoError(t, err)
			// open and read
			//got, err := str.Open(stat.Name())

			gotBytes, err := str.ReadFile(tt.args.key)
			assert.Nil(t, err)
			//defer got.Close()
			//gotBytes, err := str.ReadFile(stat.Name())
			assert.Nil(t, err)
			assert.Equal(t, tt.wantBytes, gotBytes)

			// remove
			assert.NoError(t, str.Remove(tt.args.key))
			// todo test removal of full directory path
		})
	}
}

func TestBlobStore_Remove(t *testing.T) {

	d := t.TempDir()
	fileCnt := 0
	var mu sync.Mutex
	mockFn := func(hash.Hash) string {
		mu.Lock()
		defer func() {
			fileCnt++
			mu.Unlock()
		}()

		return filepath.Join("ab", "cd", fmt.Sprintf("file-%d", fileCnt))
	}

	store, err := NewBlobStore(BlobStoreConfig{
		PathFunc: mockFn,
		Root:     d,
		Logger:   zap.Must(zap.NewDevelopment()),
	})
	require.NoError(t, err)

	f1, err := store.Create("key1")
	require.NoError(t, err)
	_, err = f1.Write([]byte("junk"))
	assert.NoError(t, err)
	assert.NoError(t, f1.Close())

	st, err := f1.Stat()
	assert.NoError(t, err)
	_, err = store.Stat(st.Name())
	assert.NoError(t, err)

	t.Logf("f name %s", st.Name())

	f2, err := store.Create("key2")
	require.NoError(t, err)
	_, err = f2.Write([]byte("other junk"))
	assert.NoError(t, err)
	assert.NoError(t, f2.Close())

	st2, err := f2.Stat()
	assert.NoError(t, err)
	_, err = store.Stat(st2.Name())
	assert.NoError(t, err)

	storePath := store.fullPath(st2.Name())
	commonDir := filepath.Dir(storePath)
	t.Logf("common dir %s", commonDir)

	dirEnts, err := store.ReadDir(commonDir)
	assert.NoError(t, err)
	assert.Len(t, dirEnts, 2)

	assert.NoError(t, store.Remove("key2"))
	_, err = store.Open("key2")
	assert.Error(t, err)

	dirEnts, err = store.ReadDir(commonDir)
	assert.NoError(t, err)
	assert.Len(t, dirEnts, 1)

	assert.NoError(t, store.Remove("key1"))
	_, err = store.Open("key1")
	assert.Error(t, err)

	dirEnts, err = store.ReadDir(commonDir)
	assert.Error(t, err)
	assert.Len(t, dirEnts, 0)

}
