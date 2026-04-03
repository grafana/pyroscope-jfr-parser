//go:build jfrparserdebug

package parser

import "fmt"

func debugSkippedEvent(typeName string) {
	fmt.Printf("skipping event type: %s\n", typeName)
}
