package tests

import("bytes")

func assertMatch(a, b []byte) bool{
    return bytes.Equal(a,b)
}
