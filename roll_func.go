package dicescript

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

func getSource() *rand.PCGSource {
	s := &rand.PCGSource{}
	s.Seed(uint64(time.Now().UnixMilli()))
	return s
}

var randSource = getSource()

func Roll(src *rand.PCGSource, dicePoints IntType, mod int) IntType {
	if dicePoints == 0 {
		return 0
	}
	// 这里判断不了IntType的长度，但编译器会自动优化掉没用的分支
	// 注: 由于 gopherJs 会因为 MaxInt64 > uint_max 而编译错误，所以限制最大值为int32，看他后续版本是否会有改进
	// if IntTypeSize == 8 && dicePoints > math.MaxInt64-1 {
	// 	return 0
	// }
	// if IntTypeSize == 4 && dicePoints > math.MaxInt32-1 {
	// 	return 0
	// }
	if dicePoints > math.MaxInt32-1 {
		return 0
	}

	if mod == -1 {
		return 1
	}
	if mod == 1 {
		return dicePoints
	}
	if src == nil {
		src = randSource
	}

	v := src.Uint64() // 如果弄32位版本，可以写成 uint32(src.Uint64() >> 32)
	n := uint64(dicePoints)
	// 下面这段取整代码来自 golang 的 exp/rand
	if n&(n-1) == 0 { // n is power of two, can mask
		return IntType(v&(n-1) + 1)
	}
	if v > math.MaxUint64-n { // Fast check.
		ceiling := math.MaxUint64 - math.MaxUint64%n
		for v >= ceiling {
			v = src.Uint64()
		}
	}
	return IntType(v%n + 1)
}

func wodCheck(e *Context, addLine IntType, pool IntType, points IntType, threshold IntType) bool {
	// makeE6 := func() {
	//	e.Error = errors.New("E6: 类型错误")
	// }

	if pool < 1 || pool > 20000 {
		e.Error = errors.New("E7: 非法数值, 骰池范围是1到20000")
		return false
	}

	if addLine != 0 && addLine < 2 {
		e.Error = errors.New("E7: 非法数值, 加骰线必须为0[不加骰]，或≥2")
		return false
	}

	if points < 1 {
		e.Error = errors.New("E7: 非法数值, 面数至少为1")
		return false
	}

	if threshold < 1 {
		e.Error = errors.New("E7: 非法数值, 成功线至少为1")
		return false
	}

	return true
}

// RollWoD 返回: 成功数，总骰数，轮数，细节
func RollWoD(src *rand.PCGSource, addLine IntType, pool IntType, points IntType, threshold IntType, isGE bool, mode int) (IntType, IntType, IntType, string) {
	var details []string
	addTimes := 1

	isShowDetails := pool < 15
	allRollCount := pool
	successCount := IntType(0)

	for times := 0; times < addTimes; times++ {
		addCount := IntType(0)
		var detailsOne []string

		for i := IntType(0); i < pool; i++ {
			var reachSuccess bool
			var reachAddRound bool
			one := Roll(src, points, mode)

			if addLine != 0 {
				reachAddRound = one >= addLine
			}

			if isGE {
				reachSuccess = one >= threshold
			} else {
				reachSuccess = one <= threshold
			}

			if reachSuccess {
				successCount += 1
			}
			if reachAddRound {
				addCount += 1
			}

			if isShowDetails {
				baseText := strconv.FormatInt(int64(one), 10)
				if reachSuccess {
					baseText += "*"
				}
				if reachAddRound {
					baseText = "<" + baseText + ">"
				}
				detailsOne = append(detailsOne, baseText)
			}
		}

		allRollCount += addCount
		// 有加骰，再骰一次
		if addCount > 0 {
			addTimes += 1
			pool = addCount
		}

		if allRollCount > 100 {
			// 多于100，清空
			isShowDetails = false
			details = details[:0]
		}

		if isShowDetails {
			details = append(details, "{"+strings.Join(detailsOne, ",")+"}")
		}
	}

	// 生成detail文本
	roundsText := ""
	if addTimes > 1 {
		roundsText = fmt.Sprintf(" 轮数:%d", addTimes)
	}

	detailText := ""
	if len(details) > 0 {
		detailText = " " + strings.Join(details, ",")
	}
	detailText = fmt.Sprintf("成功%d/%d%s%s", successCount, allRollCount, roundsText, detailText)

	// 成功数，总骰数，轮数，细节
	return successCount, allRollCount, IntType(addTimes), detailText
}

func doubleCrossCheck(ctx *Context, addLine, pool, points IntType) bool {
	if pool < 1 || pool > 20000 {
		ctx.Error = errors.New("E7: 非法数值, 骰池范围是1到20000")
		return false
	}

	if addLine < 2 {
		ctx.Error = errors.New("E7: 非法数值, 加骰线必须大于等于2")
		return false
	}

	if points < 1 {
		ctx.Error = errors.New("E7: 非法数值, 面数至少为1")
		return false
	}

	return true
}

func RollDoubleCross(src *rand.PCGSource, addLine IntType, pool IntType, points IntType, mode int) (IntType, IntType, IntType, string) {
	var details []string
	addTimes := 1

	isShowDetails := pool < 15
	allRollCount := pool
	resultDice := IntType(0)

	for times := 0; times < addTimes; times++ {
		addCount := IntType(0)
		detailsOne := []string{}
		maxDice := IntType(0)

		for i := IntType(0); i < pool; i++ {
			one := Roll(src, points, mode)
			if one > maxDice {
				maxDice = one
			}
			reachAddRound := one >= addLine

			if reachAddRound {
				addCount += 1
				maxDice = 10
			}

			if isShowDetails {
				baseText := strconv.FormatInt(int64(one), 10)
				if reachAddRound {
					baseText = "<" + baseText + ">"
				}
				detailsOne = append(detailsOne, baseText)
			}
		}

		resultDice += maxDice
		allRollCount += addCount

		// 有加骰，再骰一次
		if addCount > 0 {
			addTimes += 1
			pool = addCount
		}

		if allRollCount > 100 {
			// 多于100，清空
			isShowDetails = false
			details = details[:0]
		}

		if isShowDetails {
			details = append(details, "{"+strings.Join(detailsOne, ",")+"}")
		}
	}

	// 详细信息
	detailText := ""
	if len(details) > 0 {
		detailText = " " + strings.Join(details, ",")
	}

	roundsText := ""
	if addTimes > 1 {
		roundsText = fmt.Sprintf(" 轮数:%d", addTimes)
	}

	var lastDetail string
	if resultDice == 1 {
		lastDetail = fmt.Sprintf("大失败 出目%d/%d%s%s", resultDice, allRollCount, roundsText, detailText)
	} else {
		lastDetail = fmt.Sprintf("出目%d/%d%s%s", resultDice, allRollCount, roundsText, detailText)
	}

	// 成功数，总骰数，轮数，细节
	return resultDice, allRollCount, IntType(addTimes), lastDetail
}

// RollCommon (times)d(dicePoints)kl(lowNum) 或 (times)d(dicePoints)kh(highNum)
func RollCommon(src *rand.PCGSource, times, dicePoints IntType, diceMin, diceMax *IntType, isKeepLH, lowNum, highNum IntType, mode int) (IntType, string) {
	var nums []IntType
	for i := IntType(0); i < times; i += 1 {
		die := Roll(src, dicePoints, mode)
		if diceMax != nil {
			if die > *diceMax {
				die = *diceMax
			}
		}
		if diceMin != nil {
			if die < *diceMin {
				die = *diceMin
			}
		}
		nums = append(nums, die)
	}

	// 默认pickNum为全部，稍后由kh或kl做削减
	pickNum := times

	if isKeepLH != 0 {
		// 为1对应取低个数，为2对应取高个数，3为丢弃低个数，4为丢弃高个数
		if isKeepLH == 1 || isKeepLH == 4 {
			sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] }) // 从小到大
		} else {
			sort.Slice(nums, func(i, j int) bool { return nums[i] > nums[j] }) // 从大到小
		}

		switch isKeepLH {
		case 1, 3:
			pickNum = lowNum
		case 2, 4:
			pickNum = highNum
		}

		if isKeepLH > 2 {
			pickNum = times - pickNum
		}

		// clamp
		if pickNum < 0 {
			pickNum = 0
		}
		if pickNum > times {
			pickNum = times
		}
	}

	num := IntType(0)
	for i := IntType(0); i < pickNum; i++ {
		// 当取数大于上限 跳过
		if i >= IntType(len(nums)) {
			continue
		}
		num += nums[i]
	}

	// details
	var text string

	if pickNum == times {
		text = ""
		for i := 0; i < len(nums); i++ {
			text += fmt.Sprintf("%d+", nums[i])
		}
		if len(nums) > 0 {
			text = text[:len(text)-1]
		}
	} else {
		text = "{"
		for i := IntType(0); i < IntType(len(nums)); i++ {
			if i == pickNum {
				text += "| "
			}
			text += fmt.Sprintf("%d ", nums[i])
		}
		if len(nums) > 0 {
			text = text[:len(text)-1]
		}
		text += "}"
	}

	return num, text
}

func RollCoC(src *rand.PCGSource, isBonus bool, diceNum IntType, mode int) (IntType, string) {
	diceResult := Roll(src, 100, mode)
	diceTens := diceResult / 10
	diceUnits := diceResult % 10

	var nums []string
	diceMin := diceTens
	diceMax := diceTens
	num10Exists := false

	for i := IntType(0); i < diceNum; i++ {
		n := Roll(src, 10, mode)

		if n == 10 {
			num10Exists = true
			nums = append(nums, "0")
			continue
		} else {
			nums = append(nums, strconv.FormatInt(int64(n), 10))
		}

		if n < diceMin {
			diceMin = n
		}
		if n > diceMax {
			diceMax = n
		}
	}

	if isBonus {
		// 如果个位数不是0，那么允许十位为0
		if diceUnits != 0 && num10Exists {
			diceMin = 0
		}

		newVal := diceMin*10 + diceUnits
		lastDetail := fmt.Sprintf("(D100=%d,奖励%s)", diceResult, strings.Join(nums, " "))
		return newVal, lastDetail
	} else {
		// 如果个位数为0，那么允许十位为10
		if diceUnits == 0 && num10Exists {
			diceMax = 10
		}

		newVal := diceMax*10 + diceUnits
		lastDetail := fmt.Sprintf("(D100=%d,惩罚%s)", diceResult, strings.Join(nums, " "))
		return newVal, lastDetail
	}
}

func RollFate(src *rand.PCGSource, mode int) (IntType, string) {
	detail := ""
	sum := IntType(0)
	for i := 0; i < 4; i++ {
		n := Roll(src, 3, mode) - 2
		sum += n
		switch n {
		case -1:
			detail += "-"
		case 0:
			detail += "0"
		case +1:
			detail += "+"
		}
	}
	return sum, detail
}
