// @Time    : 2019-09-12 10:38
// @Author  : Frank
// @Email   : frank@163.com
// @File    : tools.go
// @Software: GoLand
package utils

func ContainsInt(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}