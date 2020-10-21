package example3

func myPromiseProcessor(privateData string, 
	threadState map[string]interface{}) Processor {
	return func(pr Promise) {
		me := pr.(*MyPromise)
		me.data[privateData] = 1
		if _, ok := threadState["count"]; !ok {
			threadState["count"] = []int{}
		}
		threadState["count"] = append(threadState["count"].([]int), 1)
	}
}

func myPromiseComplete(i int) Complete {
	return func(p Promise, err error) {
		me := p.(*MyPromise)
		if me.data["found"].(int) != 1 || me.ident != i {
			panic("mixed up")
		}
	}
}

// MyPromise of local data
type MyPromise struct {
	ident int
	data  map[string]interface{}
	GenericPromise
}

// MyPromiseConstructor can construct promises of type MyPromise
func MyPromiseConstructor(i int) Open {
	return func() Promise {
		p := &MyPromise{
			ident: i,
			data:  make(map[string]interface{}),
		}
		p.Super()
		return p
	}
}
