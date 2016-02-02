// Package bench is a package which contains
// programs of Go Benchmark Competition.
package bench

import (
	"syscall"
)

// Find reads the text file on the `path`,
// finds the `s` words on the file and
// returns the row numbers and indices
// in the form of `r:c,r:c,...r:c`,
// at which the `s` word exists.
func Find(path, s string) (string, error) {
	if len(s) == 0 {
		return "", syscall.EINVAL
	}
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_CLOEXEC, 0777)
	if err != nil {
		return "", err
	}

	// Target system: Linux, amd64, 1 core

	var (
		row       uint = 1
		col       uint
		needlePos uint
		i, j      uint
		lineLen   uint = 1<<32 - 1 // max uint32
		jump      bool
		needle    = []byte(s)
		lenNeedle = uint(len(needle))
		maxNeedle = lenNeedle - 1
		theByte   = needle[maxNeedle]

		// 14/2kb in/output buffer is sufficient for the benchmark sample.
		out    [20 * 1024]byte
		outLen uint
		buf    [64 * 1024]byte
		bufLen uint
	)

	// Match table.
	var match [256]bool
	match[theByte] = true
	match['\n'] = true

	col = uint(maxNeedle)

	for {
		n, err := syscall.Read(fd, buf[bufLen:])
		if err != nil {
			break
		}
		if n == 0 {
			break
		}

		bufLen += uint(n)

		for i = maxNeedle; i < bufLen; i++ {
			c := buf[i]
			if match[c] {
				if c == '\n' {
					if col < lineLen {
						lineLen = col
					}
					// We make assumptions about the data. Jump one line if we have a
					// match in the current one.
					if jump {
						col = uint(maxNeedle)
						i += lineLen - maxNeedle
						jump = false
					} else {
						col = uint(maxNeedle)
						i += maxNeedle

					}
					row++
				} else if i >= maxNeedle {
					j = i - 1
					needlePos = maxNeedle - 1
					for {
						if buf[j] != needle[needlePos] {
							break
						}
						if needlePos == 0 {
							j--
							break
						}
						j--
						needlePos--
					}

					if i-j == lenNeedle {
						jump = true

						// Write row.
						if row < 10 {
							out[outLen] = 0x30 + byte(row)
							outLen++
						} else {
							outLen += writeInt(row, out[outLen:])
						}
						out[outLen] = ':'
						outLen++

						// Write column.
						realCol := col - uint(maxNeedle)
						if realCol < 10 {
							out[outLen] = 0x30 + byte(realCol)
							outLen++
						} else {
							outLen += writeInt(realCol, out[outLen:])
						}
						out[outLen] = ','
						outLen++
					}
					col++
				}
			} else {
				col++
			}
		}

		// Move last chunk to front.
		if bufLen > 0 {
			copy(buf[:], buf[bufLen-maxNeedle:bufLen])
			bufLen = maxNeedle
		}
	}

	syscall.Close(fd)

	if outLen > 0 {
		return string(out[:outLen-1]), nil
	}

	return "", nil
}

func writeInt(n uint, b []byte) uint {
	var i uint = 0
	for {
		b[19-i] = 0x30 + byte(n%10) // 20 = max len int64
		if n < 10 {
			break
		}
		n /= 10
		i++
	}
	return uint(copy(b, b[19-i:20]))
}
