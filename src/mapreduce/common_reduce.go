package mapreduce

import (
	"io"
	"os"
	"encoding/json"
	"sort"
)

// doReduce does the job of a reduce worker: it reads the intermediate
// key/value pairs (produced by the map phase) for this task, sorts the
// intermediate key/value pairs by key, calls the user-defined reduce function
// (reduceF) for each key, and writes the output to disk.
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {

	keyToValues := make(map[string][]string)

	// Read in the content from mappers and put in correct key
	for i := 0; i < nMap; i++ {
		filename := reduceName(jobName, i, reduceTaskNumber)
		f, err := os.Open(filename)
		check_err(err)

		decoder := json.NewDecoder(f)

		for {
			var kv KeyValue
			if err := decoder.Decode(&kv); err == io.EOF {
				break
			} else if err == nil {
				keyToValues[kv.Key] = append(keyToValues[kv.Key], kv.Value)	
			}	
		}
		f.Close()
	}

	// Merge all content together under one key in sorted order
	var keys[]string 
	for k := range keyToValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	output_filename := mergeName(jobName, reduceTaskNumber)
	output_f, err := os.Create(output_filename)
	defer output_f.Close()
	check_err(err)

	encoder := json.NewEncoder(output_f)

	for _,k := range keys {
		err := encoder.Encode(KeyValue{k, reduceF(k, keyToValues[k])})
		check_err(err)
	}

}
