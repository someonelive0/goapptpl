package main

import "fmt"

func main() {
	// testdoc1()
	// testdoc2()

	// embdding_text("天空呈蓝色是因为瑞利散射。")
	// embdding_text("树叶是绿色的因为叶绿素吸收红光和蓝光。")
	contents, err := searchSimilarDocuments("为什么天空是蓝色的？", 10)
	if err != nil {
		fmt.Println(err)

	} else if len(contents) > 0 {
		fmt.Println(contents)
		CallLLM(contents[0], "为什么天空是蓝色的？")
	}

}
