package mcla

import (
	"regexp"
	"strings"
)

const ModConflictSolutionID = 12

const (
	spongepoweredInjectionErrorClass = "org.spongepowered.asm.mixin.injection.throwables.InjectionError"
)

func (a *Analyzer) HardCodedChecks(jerr *JavaError) (desc *ErrorDesc, err error) {
	if jerr.Class == spongepoweredInjectionErrorClass {
		if desc, err = a.hardCodedRedirectConflictCheck(jerr); desc != nil || err != nil {
			return
		}
	}
	return nil, nil
}

var (
	mixinRedirectConflictRe = regexp.MustCompile(`^@Redirect conflict. Skipping ([^\.]+)\.mixins\.json:[0-9A-Za-z_$]+->@Redirect::([0-9A-Za-z_$]+)\(.+already redirected by ([^\.]+)\.mixins\.json:.+`)
)

// Example:
// ```
// [16:20:56] [pool-4-thread-1/WARN] [mixin/]: @Redirect conflict. Skipping
// tfc.mixins.json:BiomeMixin->@Redirect::shouldFreezeWithClimate(Lnet/minecraft/world/level/biome/Biome;Lnet/minecraft/core/BlockPos;Lnet/minecraft/world/level/LevelReader;)Z with priority 1000,
// already redirected by
// sereneseasons.mixins.json:MixinBiome->@Redirect::onShouldFreeze_warmEnoughToRain(Lnet/minecraft/world/level/biome/Biome;Lnet/minecraft/core/BlockPos;Lnet/minecraft/world/level/LevelReader;)Z with priority 1000
// ...
// Caused by: org.spongepowered.asm.mixin.injection.throwables.InjectionError: Critical injection failure: Redirector shouldFreezeWithClimate(Lnet/minecraft/world/level/biome/Biome;Lnet/minecraft/core/BlockPos;Lnet/minecraft/world/level/LevelReader;)Z in tfc.mixins.json:BiomeMixin failed injection check, (0/1) succeeded. Scanned 1 target(s). Using refmap tfc.refmap.json
// ```
func (a *Analyzer) hardCodedRedirectConflictCheck(jerr *JavaError) (desc *ErrorDesc, err error) {
	const redirectorMessage = "Critical injection failure: Redirector "
	targetName, ok := strings.CutPrefix(jerr.Message, redirectorMessage)
	if !ok {
		return
	}
	targetName, _, ok = strings.Cut(targetName, "(")
	if !ok {
		return
	}
	var mod1, mod2, method string
	for line := range a.recentMixinLogs.IterReversed() {
		matches := mixinRedirectConflictRe.FindStringSubmatch(line)
		if matches != nil {
			mod1, method, mod2 = matches[1], matches[2], matches[3]
			break
		}
	}
	if mod1 != "" {
		return &ErrorDesc{
			Error:     spongepoweredInjectionErrorClass,
			Message:   redirectorMessage,
			Solutions: []int{ModConflictSolutionID},
			Data: map[string]any{
				"mod1":   mod1,
				"mod2":   mod2,
				"method": method,
			},
		}, nil
	}
	return
}
