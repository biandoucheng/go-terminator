package goterminator

import (
	"fmt"
)

func runFunc(fun func(), done *bool, name string) {
	defer func(flg *bool) {
		*flg = true
		if err := recover(); err != nil {
			fmt.Printf("%s: Func [%s] run failed with error %s \n.", terminatorName, name, err)
		}
	}(done)

	fun()
}

func sortFuncs(funcs map[string]int) []string {
	// Batch record by priority
	sorted_slice := map[int][]string{}
	num := len(funcs)
	for name, idx := range funcs {
		n_idx := idx % num

		names, has := sorted_slice[n_idx]
		if !has {
			names = []string{}
		}
		names = append(names, name)
		sorted_slice[n_idx] = names
	}

	// Expand by index in batches
	sorted := []string{}
	for i := 0; i < num; i++ {
		names, has := sorted_slice[i]
		if !has {
			continue
		}

		sorted = append(sorted, names...)
	}

	return sorted
}
