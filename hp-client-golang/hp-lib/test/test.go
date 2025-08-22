package main

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"time"
)

func Write(p []byte) (n int, err error) {
	limiter := rate.NewLimiter(rate.Limit(float64(1)), 1)
	b := limiter.Burst()
	for {
		end := len(p)
		if end == 0 {
			break
		}
		if b < len(p) {
			end = b
		}
		err = limiter.WaitN(context.Background(), end)
		if err != nil {
			return
		}

		fmt.Println(len(p[:end]))
		if err != nil {
			return
		}
		p = p[end:]
	}
	return
}
func main() {
	data := make([]byte, 10*1020) // 创建一个大小为10MB的字节数组

	start := time.Now()

	// 在这里插入需要统计执行时间的代码
	Write(data)

	time.Sleep(time.Second)

	duration := time.Since(start)
	fmt.Printf("代码执行时间：%s\n", duration)

}
