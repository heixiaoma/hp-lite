package main

type Info struct {
	Name string
}

func test(info Info) {
	print(info.Name)
	info.Name = "啊"
}

func main() {

	i := Info{
		Name: "解",
	}
	test(i)

	test(i)

}
