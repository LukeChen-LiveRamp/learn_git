package main

import "fmt"

/**
要求一：能够实现删除操作就可以。
要求二：考虑使用比较高性能的实现。
要求三：改造为泛型方法
要求四：支持缩容，并旦设计缩容机制。
*/

func isValidIndex[T any](slice []T, index int) bool {
	return index >= 0 && index < len(slice)
}

func RemoveIntElement(slice []int, index int) []int {
	if !isValidIndex(slice, index) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

func RemoveIntElementOptimized(slice []int, index int) []int {
	if !isValidIndex(slice, index) {
		return slice
	}
	copy(slice[index:], slice[index+1:])
	return slice[:len(slice)-1]
}

func RemoveElement[T any](slice []T, index int) []T {
	if !isValidIndex(slice, index) {
		return slice
	}
	copy(slice[index:], slice[index+1:])
	return slice[:len(slice)-1]
}

func RemoveElementWithShrink[T any](slice []T, index int) []T {
	slice = RemoveElement(slice, index)
	return ShrinkIfNeeded(slice)
}

func ShrinkIfNeeded[T any](slice []T) []T {
	const shrinkThreshold = 2
	if cap(slice) > len(slice)*shrinkThreshold {
		return append([]T(nil), slice...)
	}
	return slice
}

func main() {
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("Original slice:", nums)

	// 使用基础版本删除
	// fmt.Println("After removal:", RemoveIntElement(nums, 2))

	// 使用优化版本删除
	// fmt.Println("After removal (optimized):", RemoveIntElementOptimized(nums, 2))

	// 使用泛型版本删除
	//aab := []string{"a", "b", "c", "d", "e"}
	//fmt.Println("Original string slice:", aab)
	//fmt.Println("After removal (generics):", RemoveElement(aab, 2))

	// 使用缩容机制版本删除
	nums = RemoveElementWithShrink(nums, 0)
	fmt.Printf("After removal (shrink) first - Capacity: %d, Slice: %v\n", cap(nums), nums)
	nums = RemoveElementWithShrink(nums, 0)
	fmt.Printf("After removal (shrink) second - Capacity: %d, Slice: %v\n", cap(nums), nums)
	nums = RemoveElementWithShrink(nums, 0)
	fmt.Printf("After removal (shrink) thrid - Capacity: %d, Slice: %v\n", cap(nums), nums)
}
