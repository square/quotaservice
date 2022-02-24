// zkhelpers contains functions that make it easier to work with the primitives provided by
// github.com/samuel/go-zookeeper. This is inspired by the ZkUtil class that is included in the official Java client
// library. The eventual goal is to contribute these functions to github.com/samuel/go-zookeeper
//
// See http://grepcode.com/file/repo1.maven.org/maven2/org.apache.zookeeper/zookeeper/3.4.5/org/apache/zookeeper/ZKUtil.java
package zkhelpers

import (
	"fmt"
	"strings"

	"github.com/samuel/go-zookeeper/zk"
)

const (
	DefaultRoot                 = "/"
	internalZookeeperNode       = "/zookeeper"
	internalZookeeperNodePrefix = "/zookeeper/"
)

// ListSubtree - BFS Traversal of the system under pathRoot, with the entries in the list, in the same order as that
// of the traversal.
//
// Important: This is not an atomic snapshot of the tree ever, but the state as it exists across multiple RPCs from
// zkClient to the ensemble.
func ListSubtree(zkConn *zk.Conn, pathRoot string) ([]string, error) {
	queue := []string{pathRoot}
	tree := []string{pathRoot}
	var node string

	for {
		if len(queue) == 0 {
			// We're done
			return tree, nil
		}

		// Pop first element in the queue
		node, queue = queue[0], queue[1:]
		children, _, err := zkConn.Children(node)

		if err != nil {
			return nil, err
		}

		for _, child := range children {
			var childPath string
			if node == "/" {
				childPath = "/" + child
			} else {
				childPath = fmt.Sprintf("%v/%v", node, child)
			}
			queue = append(queue, childPath)
			tree = append(tree, childPath)
		}

	}
}

// DeleteRecursively will recursively delete the node with the given path. All versions of all nodes under the given
// node are deleted.
//
// If there is an error with deleting one of the sub-nodes in the tree, this operation would abort and would be the
// responsibility of the caller to handle the same.
func DeleteRecursively(zkConn *zk.Conn, pathRoot string) error {
	tree, err := ListSubtree(zkConn, pathRoot)
	if err != nil {
		return err
	}

	deletes := make([]interface{}, 0, len(tree))

	// We want to delete from the leaves
	for i := len(tree) - 1; i >= 0; i-- {
		if !IsInternalNode(tree[i]) && tree[i] != DefaultRoot {
			deletes = append(deletes, &zk.DeleteRequest{Path: tree[i], Version: -1})
		}
	}

	// Atomically delete all nodes
	_, err = zkConn.Multi(deletes...)
	if err != nil {
		return err
	}

	return nil
}

func IsInternalNode(path string) bool {
	return path == internalZookeeperNode || strings.HasPrefix(path, internalZookeeperNodePrefix)
}
