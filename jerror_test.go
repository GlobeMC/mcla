package mcla_test

import (
	. "github.com/GlobeMC/mcla"
	"testing"

	"strings"
)

func TestScanJavaErrors(t *testing.T) {
	const anError = `
java.lang.reflect.InvocationTargetException: null
	at jdk.internal.reflect.DirectConstructorHandleAccessor.newInstance(DirectConstructorHandleAccessor.java:74) ~[?:?]
	at java.lang.reflect.Constructor.newInstanceWithCaller(Constructor.java:502) ~[?:?]
	at java.lang.reflect.Constructor.newInstance(Constructor.java:486) ~[?:?]
	at net.minecraftforge.fml.javafmlmod.FMLModContainer.constructMod(FMLModContainer.java:67) ~[javafmllanguage-1.18.2-40.2.17.jar%23103!/:?]
	at net.minecraftforge.fml.ModContainer.lambda$buildTransitionHandler$4(ModContainer.java:122) ~[fmlcore-1.18.2-40.2.17.jar%23102!/:?]
	at java.util.concurrent.CompletableFuture$AsyncRun.run(CompletableFuture.java:1804) [?:?]
	at java.util.concurrent.CompletableFuture$AsyncRun.exec(CompletableFuture.java:1796) [?:?]
	at java.util.concurrent.ForkJoinTask.doExec(ForkJoinTask.java:387) [?:?]
	at java.util.concurrent.ForkJoinPool$WorkQueue.topLevelExec(ForkJoinPool.java:1312) [?:?]
	at java.util.concurrent.ForkJoinPool.scan(ForkJoinPool.java:1843) [?:?]
	at java.util.concurrent.ForkJoinPool.runWorker(ForkJoinPool.java:1808) [?:?]
	at java.util.concurrent.ForkJoinWorkerThread.run(ForkJoinWorkerThread.java:188) [?:?]
Caused by: java.lang.ExceptionInInitializerError
	at loaderCommon.forge.com.seibel.distanthorizons.common.wrappers.DependencySetup.createClientBindings(DependencySetup.java:69) ~[DistantHorizons-2.0.1-a-1.18.2.jar%2363!/:?]
	at com.seibel.distanthorizons.forge.ForgeMain.<init>(ForgeMain.java:98) ~[DistantHorizons-2.0.1-a-1.18.2.jar%2363!/:?]
	at jdk.internal.reflect.DirectConstructorHandleAccessor.newInstance(DirectConstructorHandleAccessor.java:62) ~[?:?]
	... 11 more
Caused by: java.lang.RuntimeException: Attempted to load class net/minecraft/client/Minecraft for invalid dist DEDICATED_SERVER
	at net.minecraftforge.fml.loading.RuntimeDistCleaner.processClassWithFlags(RuntimeDistCleaner.java:57) ~[fmlloader-1.18.2-40.2.17.jar%2318!/:1.0]
	at cpw.mods.modlauncher.LaunchPluginHandler.offerClassNodeToPlugins(LaunchPluginHandler.java:88) ~[modlauncher-9.1.3.jar%235!/:?]
	at cpw.mods.modlauncher.ClassTransformer.transform(ClassTransformer.java:120) ~[modlauncher-9.1.3.jar%235!/:?]
	at cpw.mods.modlauncher.TransformingClassLoader.maybeTransformClassBytes(TransformingClassLoader.java:50) ~[modlauncher-9.1.3.jar%235!/:?]
	at cpw.mods.cl.ModuleClassLoader.readerToClass(ModuleClassLoader.java:113) ~[securejarhandler-1.0.8.jar:?]
	at cpw.mods.cl.ModuleClassLoader.lambda$findClass$15(ModuleClassLoader.java:219) ~[securejarhandler-1.0.8.jar:?]
	at cpw.mods.cl.ModuleClassLoader.loadFromModule(ModuleClassLoader.java:229) ~[securejarhandler-1.0.8.jar:?]
	at cpw.mods.cl.ModuleClassLoader.findClass(ModuleClassLoader.java:219) ~[securejarhandler-1.0.8.jar:?]
	at cpw.mods.cl.ModuleClassLoader.loadClass(ModuleClassLoader.java:135) ~[securejarhandler-1.0.8.jar:?]
	at java.lang.ClassLoader.loadClass(ClassLoader.java:526) ~[?:?]
	at loaderCommon.forge.com.seibel.distanthorizons.common.wrappers.minecraft.MinecraftClientWrapper.<init>(MinecraftClientWrapper.java:71) ~[DistantHorizons-2.0.1-a-1.18.2.jar%2363!/:?]
	at loaderCommon.forge.com.seibel.distanthorizons.common.wrappers.minecraft.MinecraftClientWrapper.<clinit>(MinecraftClientWrapper.java:69) ~[DistantHorizons-2.0.1-a-1.18.2.jar%2363!/:?]
	at loaderCommon.forge.com.seibel.distanthorizons.common.wrappers.DependencySetup.createClientBindings(DependencySetup.java:69) ~[DistantHorizons-2.0.1-a-1.18.2.jar%2363!/:?]
	at com.seibel.distanthorizons.forge.ForgeMain.<init>(ForgeMain.java:98) ~[DistantHorizons-2.0.1-a-1.18.2.jar%2363!/:?]
	at jdk.internal.reflect.DirectConstructorHandleAccessor.newInstance(DirectConstructorHandleAccessor.java:62) ~[?:?]
	... 11 more

Caused by: java.lang.RuntimeException:
	at a.b.c.d.E.f
`

	res, err := ScanJavaErrors(strings.NewReader(anError))
	if err != nil {
		t.Errorf("Cannot parse anError: %v", err)
		return
	}
	if len(res) != 1 {
		t.Errorf("Found %d java errors, but expect only 1", len(res))
		return
	}
	je := res[0]
	if expect := "java.lang.reflect.InvocationTargetException"; je.Class != expect {
		t.Errorf(`Expect je.Class == %q, got %q`, expect, je.Class)
		return
	}
	if expect := "null"; je.Message != expect {
		t.Errorf(`Expect je.Message == %q, got %q`, expect, je.Message)
		return
	}
	if je.CausedBy == nil {
		t.Errorf(`Expect je.CausedBy != nil, got nil`)
		return
	}
	je = je.CausedBy
	if expect := "java.lang.ExceptionInInitializerError"; je.Class != expect {
		t.Errorf(`Expect je.Class == %q, got %q`, expect, je.Class)
		return
	}
	if expect := ""; je.Message != expect {
		t.Errorf(`Expect je.Message == %q, got %q`, expect, je.Message)
		return
	}
	if je.CausedBy == nil {
		t.Errorf(`Expect je.CausedBy != nil, got nil`)
		return
	}
	je = je.CausedBy
	if expect := "java.lang.RuntimeException"; je.Class != expect {
		t.Errorf(`Expect je.Class == %q, got %q`, expect, je.Class)
		return
	}
	if expect := "Attempted to load class net/minecraft/client/Minecraft for invalid dist DEDICATED_SERVER"; je.Message != expect {
		t.Errorf(`Expect je.Message == %q, got %q`, expect, je.Message)
		return
	}
	if je.CausedBy != nil {
		t.Errorf(`Expect je.CausedBy == nil, got %#v`, je.CausedBy)
		return
	}
}
