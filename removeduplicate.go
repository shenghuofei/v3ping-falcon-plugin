package main

func RemoveDuplicate(list []string) []string {
	var x []string = []string{}
	for _, i := range list {
		if len(x) == 0 {
			x = append(x, i)
		} else {
			for k, v := range x {
				if i == v {
					break
				}
				if k == len(x)-1 {
					x = append(x, i)
				}
			}
		}
	}
	return x
}

func RemoveDuplicate2(list []string) []string {
	slen := len(list)
	var j int
	var x []string = []string{}
	for i, v := range list {
		for j = i + 1; j < slen; j++ {
			if v == list[j] {
				break
			}
		}
		if j == slen {
			x = append(x, v)
		}
	}
	return x
}
