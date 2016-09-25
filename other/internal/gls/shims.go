// +build !goid

package gls

import (
	"reflect"
	"runtime"
)

func init() {
	basePtr := reflect.ValueOf(shim00).Pointer()
	var pc uintptr
	f := func(_ []shim) {
		pcs := make([]uintptr, 8)
		runtime.Callers(2, pcs)
		pc = pcs[0]
	}
	shim00([]shim{f})
	shimPCOffset = basePtr - pc

	for i, f := range shims {
		var pc uintptr
		g := func(_ []shim) {
			pcs := make([]uintptr, 8)
			runtime.Callers(2, pcs)
			pc = pcs[0]
		}
		f([]shim{g})
		pcToUintptr[pc] = uintptr(i)
	}
}

var shims = []shim{shim00, shim01, shim02, shim03, shim04,
	shim05, shim06, shim07, shim08, shim09, shim0A, shim0B,
	shim0C, shim0D, shim0E, shim0F, shim10, shim11, shim12,
	shim13, shim14, shim15, shim16, shim17, shim18, shim19,
	shim1A, shim1B, shim1C, shim1D, shim1E, shim1F, shim20,
	shim21, shim22, shim23, shim24, shim25, shim26, shim27,
	shim28, shim29, shim2A, shim2B, shim2C, shim2D, shim2E,
	shim2F, shim30, shim31, shim32, shim33, shim34, shim35,
	shim36, shim37, shim38, shim39, shim3A, shim3B, shim3C,
	shim3D, shim3E, shim3F, shim40, shim41, shim42, shim43,
	shim44, shim45, shim46, shim47, shim48, shim49, shim4A,
	shim4B, shim4C, shim4D, shim4E, shim4F, shim50, shim51,
	shim52, shim53, shim54, shim55, shim56, shim57, shim58,
	shim59, shim5A, shim5B, shim5C, shim5D, shim5E, shim5F,
	shim60, shim61, shim62, shim63, shim64, shim65, shim66,
	shim67, shim68, shim69, shim6A, shim6B, shim6C, shim6D,
	shim6E, shim6F, shim70, shim71, shim72, shim73, shim74,
	shim75, shim76, shim77, shim78, shim79, shim7A, shim7B,
	shim7C, shim7D, shim7E, shim7F, shim80, shim81, shim82,
	shim83, shim84, shim85, shim86, shim87, shim88, shim89,
	shim8A, shim8B, shim8C, shim8D, shim8E, shim8F, shim90,
	shim91, shim92, shim93, shim94, shim95, shim96, shim97,
	shim98, shim99, shim9A, shim9B, shim9C, shim9D, shim9E,
	shim9F, shimA0, shimA1, shimA2, shimA3, shimA4, shimA5,
	shimA6, shimA7, shimA8, shimA9, shimAA, shimAB, shimAC,
	shimAD, shimAE, shimAF, shimB0, shimB1, shimB2, shimB3,
	shimB4, shimB5, shimB6, shimB7, shimB8, shimB9, shimBA,
	shimBB, shimBC, shimBD, shimBE, shimBF, shimC0, shimC1,
	shimC2, shimC3, shimC4, shimC5, shimC6, shimC7, shimC8,
	shimC9, shimCA, shimCB, shimCC, shimCD, shimCE, shimCF,
	shimD0, shimD1, shimD2, shimD3, shimD4, shimD5, shimD6,
	shimD7, shimD8, shimD9, shimDA, shimDB, shimDC, shimDD,
	shimDE, shimDF, shimE0, shimE1, shimE2, shimE3, shimE4,
	shimE5, shimE6, shimE7, shimE8, shimE9, shimEA, shimEB,
	shimEC, shimED, shimEE, shimEF, shimF0, shimF1, shimF2,
	shimF3, shimF4, shimF5, shimF6, shimF7, shimF8, shimF9,
	shimFA, shimFB, shimFC, shimFD, shimFE, shimFF}

func shim00(shims []shim) { shims[0](shims[1:]) }
func shim01(shims []shim) { shims[0](shims[1:]) }
func shim02(shims []shim) { shims[0](shims[1:]) }
func shim03(shims []shim) { shims[0](shims[1:]) }
func shim04(shims []shim) { shims[0](shims[1:]) }
func shim05(shims []shim) { shims[0](shims[1:]) }
func shim06(shims []shim) { shims[0](shims[1:]) }
func shim07(shims []shim) { shims[0](shims[1:]) }
func shim08(shims []shim) { shims[0](shims[1:]) }
func shim09(shims []shim) { shims[0](shims[1:]) }
func shim0A(shims []shim) { shims[0](shims[1:]) }
func shim0B(shims []shim) { shims[0](shims[1:]) }
func shim0C(shims []shim) { shims[0](shims[1:]) }
func shim0D(shims []shim) { shims[0](shims[1:]) }
func shim0E(shims []shim) { shims[0](shims[1:]) }
func shim0F(shims []shim) { shims[0](shims[1:]) }
func shim10(shims []shim) { shims[0](shims[1:]) }
func shim11(shims []shim) { shims[0](shims[1:]) }
func shim12(shims []shim) { shims[0](shims[1:]) }
func shim13(shims []shim) { shims[0](shims[1:]) }
func shim14(shims []shim) { shims[0](shims[1:]) }
func shim15(shims []shim) { shims[0](shims[1:]) }
func shim16(shims []shim) { shims[0](shims[1:]) }
func shim17(shims []shim) { shims[0](shims[1:]) }
func shim18(shims []shim) { shims[0](shims[1:]) }
func shim19(shims []shim) { shims[0](shims[1:]) }
func shim1A(shims []shim) { shims[0](shims[1:]) }
func shim1B(shims []shim) { shims[0](shims[1:]) }
func shim1C(shims []shim) { shims[0](shims[1:]) }
func shim1D(shims []shim) { shims[0](shims[1:]) }
func shim1E(shims []shim) { shims[0](shims[1:]) }
func shim1F(shims []shim) { shims[0](shims[1:]) }
func shim20(shims []shim) { shims[0](shims[1:]) }
func shim21(shims []shim) { shims[0](shims[1:]) }
func shim22(shims []shim) { shims[0](shims[1:]) }
func shim23(shims []shim) { shims[0](shims[1:]) }
func shim24(shims []shim) { shims[0](shims[1:]) }
func shim25(shims []shim) { shims[0](shims[1:]) }
func shim26(shims []shim) { shims[0](shims[1:]) }
func shim27(shims []shim) { shims[0](shims[1:]) }
func shim28(shims []shim) { shims[0](shims[1:]) }
func shim29(shims []shim) { shims[0](shims[1:]) }
func shim2A(shims []shim) { shims[0](shims[1:]) }
func shim2B(shims []shim) { shims[0](shims[1:]) }
func shim2C(shims []shim) { shims[0](shims[1:]) }
func shim2D(shims []shim) { shims[0](shims[1:]) }
func shim2E(shims []shim) { shims[0](shims[1:]) }
func shim2F(shims []shim) { shims[0](shims[1:]) }
func shim30(shims []shim) { shims[0](shims[1:]) }
func shim31(shims []shim) { shims[0](shims[1:]) }
func shim32(shims []shim) { shims[0](shims[1:]) }
func shim33(shims []shim) { shims[0](shims[1:]) }
func shim34(shims []shim) { shims[0](shims[1:]) }
func shim35(shims []shim) { shims[0](shims[1:]) }
func shim36(shims []shim) { shims[0](shims[1:]) }
func shim37(shims []shim) { shims[0](shims[1:]) }
func shim38(shims []shim) { shims[0](shims[1:]) }
func shim39(shims []shim) { shims[0](shims[1:]) }
func shim3A(shims []shim) { shims[0](shims[1:]) }
func shim3B(shims []shim) { shims[0](shims[1:]) }
func shim3C(shims []shim) { shims[0](shims[1:]) }
func shim3D(shims []shim) { shims[0](shims[1:]) }
func shim3E(shims []shim) { shims[0](shims[1:]) }
func shim3F(shims []shim) { shims[0](shims[1:]) }
func shim40(shims []shim) { shims[0](shims[1:]) }
func shim41(shims []shim) { shims[0](shims[1:]) }
func shim42(shims []shim) { shims[0](shims[1:]) }
func shim43(shims []shim) { shims[0](shims[1:]) }
func shim44(shims []shim) { shims[0](shims[1:]) }
func shim45(shims []shim) { shims[0](shims[1:]) }
func shim46(shims []shim) { shims[0](shims[1:]) }
func shim47(shims []shim) { shims[0](shims[1:]) }
func shim48(shims []shim) { shims[0](shims[1:]) }
func shim49(shims []shim) { shims[0](shims[1:]) }
func shim4A(shims []shim) { shims[0](shims[1:]) }
func shim4B(shims []shim) { shims[0](shims[1:]) }
func shim4C(shims []shim) { shims[0](shims[1:]) }
func shim4D(shims []shim) { shims[0](shims[1:]) }
func shim4E(shims []shim) { shims[0](shims[1:]) }
func shim4F(shims []shim) { shims[0](shims[1:]) }
func shim50(shims []shim) { shims[0](shims[1:]) }
func shim51(shims []shim) { shims[0](shims[1:]) }
func shim52(shims []shim) { shims[0](shims[1:]) }
func shim53(shims []shim) { shims[0](shims[1:]) }
func shim54(shims []shim) { shims[0](shims[1:]) }
func shim55(shims []shim) { shims[0](shims[1:]) }
func shim56(shims []shim) { shims[0](shims[1:]) }
func shim57(shims []shim) { shims[0](shims[1:]) }
func shim58(shims []shim) { shims[0](shims[1:]) }
func shim59(shims []shim) { shims[0](shims[1:]) }
func shim5A(shims []shim) { shims[0](shims[1:]) }
func shim5B(shims []shim) { shims[0](shims[1:]) }
func shim5C(shims []shim) { shims[0](shims[1:]) }
func shim5D(shims []shim) { shims[0](shims[1:]) }
func shim5E(shims []shim) { shims[0](shims[1:]) }
func shim5F(shims []shim) { shims[0](shims[1:]) }
func shim60(shims []shim) { shims[0](shims[1:]) }
func shim61(shims []shim) { shims[0](shims[1:]) }
func shim62(shims []shim) { shims[0](shims[1:]) }
func shim63(shims []shim) { shims[0](shims[1:]) }
func shim64(shims []shim) { shims[0](shims[1:]) }
func shim65(shims []shim) { shims[0](shims[1:]) }
func shim66(shims []shim) { shims[0](shims[1:]) }
func shim67(shims []shim) { shims[0](shims[1:]) }
func shim68(shims []shim) { shims[0](shims[1:]) }
func shim69(shims []shim) { shims[0](shims[1:]) }
func shim6A(shims []shim) { shims[0](shims[1:]) }
func shim6B(shims []shim) { shims[0](shims[1:]) }
func shim6C(shims []shim) { shims[0](shims[1:]) }
func shim6D(shims []shim) { shims[0](shims[1:]) }
func shim6E(shims []shim) { shims[0](shims[1:]) }
func shim6F(shims []shim) { shims[0](shims[1:]) }
func shim70(shims []shim) { shims[0](shims[1:]) }
func shim71(shims []shim) { shims[0](shims[1:]) }
func shim72(shims []shim) { shims[0](shims[1:]) }
func shim73(shims []shim) { shims[0](shims[1:]) }
func shim74(shims []shim) { shims[0](shims[1:]) }
func shim75(shims []shim) { shims[0](shims[1:]) }
func shim76(shims []shim) { shims[0](shims[1:]) }
func shim77(shims []shim) { shims[0](shims[1:]) }
func shim78(shims []shim) { shims[0](shims[1:]) }
func shim79(shims []shim) { shims[0](shims[1:]) }
func shim7A(shims []shim) { shims[0](shims[1:]) }
func shim7B(shims []shim) { shims[0](shims[1:]) }
func shim7C(shims []shim) { shims[0](shims[1:]) }
func shim7D(shims []shim) { shims[0](shims[1:]) }
func shim7E(shims []shim) { shims[0](shims[1:]) }
func shim7F(shims []shim) { shims[0](shims[1:]) }
func shim80(shims []shim) { shims[0](shims[1:]) }
func shim81(shims []shim) { shims[0](shims[1:]) }
func shim82(shims []shim) { shims[0](shims[1:]) }
func shim83(shims []shim) { shims[0](shims[1:]) }
func shim84(shims []shim) { shims[0](shims[1:]) }
func shim85(shims []shim) { shims[0](shims[1:]) }
func shim86(shims []shim) { shims[0](shims[1:]) }
func shim87(shims []shim) { shims[0](shims[1:]) }
func shim88(shims []shim) { shims[0](shims[1:]) }
func shim89(shims []shim) { shims[0](shims[1:]) }
func shim8A(shims []shim) { shims[0](shims[1:]) }
func shim8B(shims []shim) { shims[0](shims[1:]) }
func shim8C(shims []shim) { shims[0](shims[1:]) }
func shim8D(shims []shim) { shims[0](shims[1:]) }
func shim8E(shims []shim) { shims[0](shims[1:]) }
func shim8F(shims []shim) { shims[0](shims[1:]) }
func shim90(shims []shim) { shims[0](shims[1:]) }
func shim91(shims []shim) { shims[0](shims[1:]) }
func shim92(shims []shim) { shims[0](shims[1:]) }
func shim93(shims []shim) { shims[0](shims[1:]) }
func shim94(shims []shim) { shims[0](shims[1:]) }
func shim95(shims []shim) { shims[0](shims[1:]) }
func shim96(shims []shim) { shims[0](shims[1:]) }
func shim97(shims []shim) { shims[0](shims[1:]) }
func shim98(shims []shim) { shims[0](shims[1:]) }
func shim99(shims []shim) { shims[0](shims[1:]) }
func shim9A(shims []shim) { shims[0](shims[1:]) }
func shim9B(shims []shim) { shims[0](shims[1:]) }
func shim9C(shims []shim) { shims[0](shims[1:]) }
func shim9D(shims []shim) { shims[0](shims[1:]) }
func shim9E(shims []shim) { shims[0](shims[1:]) }
func shim9F(shims []shim) { shims[0](shims[1:]) }
func shimA0(shims []shim) { shims[0](shims[1:]) }
func shimA1(shims []shim) { shims[0](shims[1:]) }
func shimA2(shims []shim) { shims[0](shims[1:]) }
func shimA3(shims []shim) { shims[0](shims[1:]) }
func shimA4(shims []shim) { shims[0](shims[1:]) }
func shimA5(shims []shim) { shims[0](shims[1:]) }
func shimA6(shims []shim) { shims[0](shims[1:]) }
func shimA7(shims []shim) { shims[0](shims[1:]) }
func shimA8(shims []shim) { shims[0](shims[1:]) }
func shimA9(shims []shim) { shims[0](shims[1:]) }
func shimAA(shims []shim) { shims[0](shims[1:]) }
func shimAB(shims []shim) { shims[0](shims[1:]) }
func shimAC(shims []shim) { shims[0](shims[1:]) }
func shimAD(shims []shim) { shims[0](shims[1:]) }
func shimAE(shims []shim) { shims[0](shims[1:]) }
func shimAF(shims []shim) { shims[0](shims[1:]) }
func shimB0(shims []shim) { shims[0](shims[1:]) }
func shimB1(shims []shim) { shims[0](shims[1:]) }
func shimB2(shims []shim) { shims[0](shims[1:]) }
func shimB3(shims []shim) { shims[0](shims[1:]) }
func shimB4(shims []shim) { shims[0](shims[1:]) }
func shimB5(shims []shim) { shims[0](shims[1:]) }
func shimB6(shims []shim) { shims[0](shims[1:]) }
func shimB7(shims []shim) { shims[0](shims[1:]) }
func shimB8(shims []shim) { shims[0](shims[1:]) }
func shimB9(shims []shim) { shims[0](shims[1:]) }
func shimBA(shims []shim) { shims[0](shims[1:]) }
func shimBB(shims []shim) { shims[0](shims[1:]) }
func shimBC(shims []shim) { shims[0](shims[1:]) }
func shimBD(shims []shim) { shims[0](shims[1:]) }
func shimBE(shims []shim) { shims[0](shims[1:]) }
func shimBF(shims []shim) { shims[0](shims[1:]) }
func shimC0(shims []shim) { shims[0](shims[1:]) }
func shimC1(shims []shim) { shims[0](shims[1:]) }
func shimC2(shims []shim) { shims[0](shims[1:]) }
func shimC3(shims []shim) { shims[0](shims[1:]) }
func shimC4(shims []shim) { shims[0](shims[1:]) }
func shimC5(shims []shim) { shims[0](shims[1:]) }
func shimC6(shims []shim) { shims[0](shims[1:]) }
func shimC7(shims []shim) { shims[0](shims[1:]) }
func shimC8(shims []shim) { shims[0](shims[1:]) }
func shimC9(shims []shim) { shims[0](shims[1:]) }
func shimCA(shims []shim) { shims[0](shims[1:]) }
func shimCB(shims []shim) { shims[0](shims[1:]) }
func shimCC(shims []shim) { shims[0](shims[1:]) }
func shimCD(shims []shim) { shims[0](shims[1:]) }
func shimCE(shims []shim) { shims[0](shims[1:]) }
func shimCF(shims []shim) { shims[0](shims[1:]) }
func shimD0(shims []shim) { shims[0](shims[1:]) }
func shimD1(shims []shim) { shims[0](shims[1:]) }
func shimD2(shims []shim) { shims[0](shims[1:]) }
func shimD3(shims []shim) { shims[0](shims[1:]) }
func shimD4(shims []shim) { shims[0](shims[1:]) }
func shimD5(shims []shim) { shims[0](shims[1:]) }
func shimD6(shims []shim) { shims[0](shims[1:]) }
func shimD7(shims []shim) { shims[0](shims[1:]) }
func shimD8(shims []shim) { shims[0](shims[1:]) }
func shimD9(shims []shim) { shims[0](shims[1:]) }
func shimDA(shims []shim) { shims[0](shims[1:]) }
func shimDB(shims []shim) { shims[0](shims[1:]) }
func shimDC(shims []shim) { shims[0](shims[1:]) }
func shimDD(shims []shim) { shims[0](shims[1:]) }
func shimDE(shims []shim) { shims[0](shims[1:]) }
func shimDF(shims []shim) { shims[0](shims[1:]) }
func shimE0(shims []shim) { shims[0](shims[1:]) }
func shimE1(shims []shim) { shims[0](shims[1:]) }
func shimE2(shims []shim) { shims[0](shims[1:]) }
func shimE3(shims []shim) { shims[0](shims[1:]) }
func shimE4(shims []shim) { shims[0](shims[1:]) }
func shimE5(shims []shim) { shims[0](shims[1:]) }
func shimE6(shims []shim) { shims[0](shims[1:]) }
func shimE7(shims []shim) { shims[0](shims[1:]) }
func shimE8(shims []shim) { shims[0](shims[1:]) }
func shimE9(shims []shim) { shims[0](shims[1:]) }
func shimEA(shims []shim) { shims[0](shims[1:]) }
func shimEB(shims []shim) { shims[0](shims[1:]) }
func shimEC(shims []shim) { shims[0](shims[1:]) }
func shimED(shims []shim) { shims[0](shims[1:]) }
func shimEE(shims []shim) { shims[0](shims[1:]) }
func shimEF(shims []shim) { shims[0](shims[1:]) }
func shimF0(shims []shim) { shims[0](shims[1:]) }
func shimF1(shims []shim) { shims[0](shims[1:]) }
func shimF2(shims []shim) { shims[0](shims[1:]) }
func shimF3(shims []shim) { shims[0](shims[1:]) }
func shimF4(shims []shim) { shims[0](shims[1:]) }
func shimF5(shims []shim) { shims[0](shims[1:]) }
func shimF6(shims []shim) { shims[0](shims[1:]) }
func shimF7(shims []shim) { shims[0](shims[1:]) }
func shimF8(shims []shim) { shims[0](shims[1:]) }
func shimF9(shims []shim) { shims[0](shims[1:]) }
func shimFA(shims []shim) { shims[0](shims[1:]) }
func shimFB(shims []shim) { shims[0](shims[1:]) }
func shimFC(shims []shim) { shims[0](shims[1:]) }
func shimFD(shims []shim) { shims[0](shims[1:]) }
func shimFE(shims []shim) { shims[0](shims[1:]) }
func shimFF(shims []shim) { shims[0](shims[1:]) }
