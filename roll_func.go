package dicescript

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

func Roll(dicePoints int64) int64 {
	if dicePoints == 0 {
		return 0
	}
	val := rand.Int63()%dicePoints + 1
	return val
}

func wodCheck(e *Context, addLine int64, pool int64, points int64, threshold int64) bool {
	//makeE6 := func() {
	//	e.Error = errors.New("E6: 类型错误")
	//}

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
func RollWoD(addLine int64, pool int64, points int64, threshold int64, isGE bool) (int64, int64, int64, string) {
	var details []string
	addTimes := 1

	isShowDetails := pool < 15
	allRollCount := pool
	successCount := int64(0)

	for times := 0; times < addTimes; times++ {
		addCount := int64(0)
		detailsOne := []string{}

		for i := int64(0); i < pool; i++ {
			var reachSuccess bool
			var reachAddRound bool
			one := Roll(points)

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
				baseText := strconv.FormatInt(one, 10)
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
	return successCount, allRollCount, int64(addTimes), detailText
}

func doubleCrossCheck(ctx *Context, addLine, pool, points int64) bool {
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

func RollDoubleCross(addLine int64, pool int64, points int64) (int64, int64, int64, string) {
	var details []string
	addTimes := 1

	isShowDetails := pool < 15
	allRollCount := pool
	resultDice := int64(0)

	for times := 0; times < addTimes; times++ {
		addCount := int64(0)
		detailsOne := []string{}
		maxDice := int64(0)

		for i := int64(0); i < pool; i++ {
			one := Roll(points)
			if one > maxDice {
				maxDice = one
			}
			reachAddRound := one >= addLine

			if reachAddRound {
				addCount += 1
				maxDice = 10
			}

			if isShowDetails {
				baseText := strconv.FormatInt(one, 10)
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
	return resultDice, allRollCount, int64(addTimes), lastDetail
}

// RollCommon (times)d(dicePoints)kl(lowNum) 或 (times)d(dicePoints)kh(highNum)
func RollCommon(times, dicePoints int64, diceMin, diceMax *int64, isKeepLH, lowNum, highNum int64) (int64, string) {
	var nums []int64
	for i := int64(0); i < times; i += 1 {
		die := Roll(dicePoints)
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
		if isKeepLH == 1 || isKeepLH == 3 {
			pickNum = lowNum
			sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] }) // 从小到大
		} else {
			pickNum = highNum
			sort.Slice(nums, func(i, j int) bool { return nums[i] > nums[j] }) // 从大到小
		}
		if isKeepLH > 2 {
			pickNum = times - pickNum
		}
	}

	num := int64(0)
	for i := int64(0); i < pickNum; i++ {
		// 当取数大于上限 跳过
		if i >= int64(len(nums)) {
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
		for i := int64(0); i < int64(len(nums)); i++ {
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

func RollCoC(isBonus bool, diceNum int64) (int64, string) {
	diceResult := Roll(100)
	diceTens := diceResult / 10
	diceUnits := diceResult % 10

	nums := []string{}
	diceMin := diceTens
	diceMax := diceTens
	num10Exists := false

	for i := int64(0); i < diceNum; i++ {
		n := Roll(10)

		if n == 10 {
			num10Exists = true
			nums = append(nums, "0")
			continue
		} else {
			nums = append(nums, strconv.FormatInt(n, 10))
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
