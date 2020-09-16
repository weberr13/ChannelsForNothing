package example3

// MyPromise of local data
type MyPromise struct {
	ident int
	data  map[string]interface{}
	GenericPromise
}

// MyPromiseConstructor can construct promises of type MyPromise
func MyPromiseConstructor(i int) func() Promise {
	return func() Promise {
		p := &MyPromise{
			ident: i,
			data:  make(map[string]interface{}),
		}
		p.Init()
		return p
	}
}
