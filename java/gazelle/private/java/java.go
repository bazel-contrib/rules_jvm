package java

import (
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

// IsTestPackage tries to detect if the directory would contain test files of not.
// It assumes dir is a forward-slashed package name, not a possibly-back-slashed filepath.
func IsTestPackage(pkg string) bool {
	if strings.HasPrefix(pkg, "javatests/") {
		return true
	}

	if strings.Contains(pkg, "src/") {
		afterSrc := strings.SplitAfterN(pkg, "src/", 2)[1]
		firstDir := strings.Split(afterSrc, "/")[0]
		if strings.HasSuffix(strings.ToLower(firstDir), "test") {
			return true
		}
	}

	return strings.Contains(pkg, "/test/")
}

// This list was derived from a script along the lines of:
// for jmod in ${JAVA_HOME}/jmods/*; do unzip -l "${jmod}" 2>/dev/null; done | grep classes/ | awk '{print $4}' | sed -e 's#^classes/##' -e 's#\.class$##' | xargs -n1 dirname | sort | uniq | sed -e 's#/#.#g'
var stdlibPrefixes = []types.PackageName{
	types.NewPackageName("com.sun.accessibility.internal.resources"),
	types.NewPackageName("com.sun.beans"),
	types.NewPackageName("com.sun.crypto.provider"),
	types.NewPackageName("com.sun.imageio.plugins.bmp"),
	types.NewPackageName("com.sun.imageio.plugins.common"),
	types.NewPackageName("com.sun.imageio.plugins.gif"),
	types.NewPackageName("com.sun.imageio.plugins.jpeg"),
	types.NewPackageName("com.sun.imageio.plugins.png"),
	types.NewPackageName("com.sun.imageio.plugins.tiff"),
	types.NewPackageName("com.sun.imageio.plugins.wbmp"),
	types.NewPackageName("com.sun.imageio.spi"),
	types.NewPackageName("com.sun.imageio.stream"),
	types.NewPackageName("com.sun.jarsigner"),
	types.NewPackageName("com.sun.java.accessibility.util"),
	types.NewPackageName("com.sun.java.swing"),
	types.NewPackageName("com.sun.java_cup.internal.runtime"),
	types.NewPackageName("com.sun.jdi"),
	types.NewPackageName("com.sun.jmx.defaults"),
	types.NewPackageName("com.sun.jmx.interceptor"),
	types.NewPackageName("com.sun.jmx.mbeanserver"),
	types.NewPackageName("com.sun.jmx.remote.internal"),
	types.NewPackageName("com.sun.jmx.remote.protocol.rmi"),
	types.NewPackageName("com.sun.jmx.remote.security"),
	types.NewPackageName("com.sun.jmx.remote.util"),
	types.NewPackageName("com.sun.jndi.dns"),
	types.NewPackageName("com.sun.jndi.ldap"),
	types.NewPackageName("com.sun.jndi.rmi.registry"),
	types.NewPackageName("com.sun.jndi.toolkit.ctx"),
	types.NewPackageName("com.sun.jndi.toolkit.dir"),
	types.NewPackageName("com.sun.jndi.toolkit.url"),
	types.NewPackageName("com.sun.jndi.url.dns"),
	types.NewPackageName("com.sun.jndi.url.ldap"),
	types.NewPackageName("com.sun.jndi.url.ldaps"),
	types.NewPackageName("com.sun.jndi.url.rmi"),
	types.NewPackageName("com.sun.management"),
	types.NewPackageName("com.sun.media.sound"),
	types.NewPackageName("com.sun.naming.internal"),
	types.NewPackageName("com.sun.net.httpserver"),
	types.NewPackageName("com.sun.nio.file"),
	types.NewPackageName("com.sun.nio.sctp"),
	types.NewPackageName("com.sun.org.apache.bcel.internal"),
	types.NewPackageName("com.sun.org.apache.xalan.internal.extensions"),
	types.NewPackageName("com.sun.org.apache.xalan.internal.lib"),
	types.NewPackageName("com.sun.org.apache.xalan.internal.res"),
	types.NewPackageName("com.sun.org.apache.xalan.internal.templates"),
	types.NewPackageName("com.sun.org.apache.xalan.internal.utils"),
	types.NewPackageName("com.sun.org.apache.xalan.internal.xsltc"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.dom"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.impl"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.jaxp"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.parsers"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.util"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.utils"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.xinclude"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.xni"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.xpointer"),
	types.NewPackageName("com.sun.org.apache.xerces.internal.xs"),
	types.NewPackageName("com.sun.org.apache.xml.internal.dtm"),
	types.NewPackageName("com.sun.org.apache.xml.internal.res"),
	types.NewPackageName("com.sun.org.apache.xml.internal.security"),
	types.NewPackageName("com.sun.org.apache.xml.internal.serialize"),
	types.NewPackageName("com.sun.org.apache.xml.internal.serializer"),
	types.NewPackageName("com.sun.org.apache.xml.internal.utils"),
	types.NewPackageName("com.sun.org.apache.xpath.internal"),
	types.NewPackageName("com.sun.org.slf4j.internal"),
	types.NewPackageName("com.sun.rowset"),
	types.NewPackageName("com.sun.security.auth"),
	types.NewPackageName("com.sun.security.jgss"),
	types.NewPackageName("com.sun.security.ntlm"),
	types.NewPackageName("com.sun.security.sasl"),
	types.NewPackageName("com.sun.source.doctree"),
	types.NewPackageName("com.sun.source.tree"),
	types.NewPackageName("com.sun.source.util"),
	types.NewPackageName("com.sun.swing.internal.plaf.basic.resources"),
	types.NewPackageName("com.sun.swing.internal.plaf.metal.resources"),
	types.NewPackageName("com.sun.swing.internal.plaf.synth.resources"),
	types.NewPackageName("com.sun.tools.attach"),
	types.NewPackageName("com.sun.tools.classfile"),
	types.NewPackageName("com.sun.tools.doclint"),
	types.NewPackageName("com.sun.tools.example.debug.expr"),
	types.NewPackageName("com.sun.tools.example.debug.tty"),
	types.NewPackageName("com.sun.tools.javac"),
	types.NewPackageName("com.sun.tools.javap"),
	types.NewPackageName("com.sun.tools.jconsole"),
	types.NewPackageName("com.sun.tools.jdeprscan"),
	types.NewPackageName("com.sun.tools.jdeps"),
	types.NewPackageName("com.sun.tools.jdi"),
	types.NewPackageName("com.sun.tools.script.shell"),
	types.NewPackageName("com.sun.tools.sjavac"),
	types.NewPackageName("com.sun.xml.internal.stream"),
	types.NewPackageName("java"),
	types.NewPackageName("javax.accessibility"),
	types.NewPackageName("javax.annotation.processing"),
	types.NewPackageName("javax.annotation.security"),
	types.NewPackageName("javax.crypto"),
	types.NewPackageName("javax.imageio"),
	types.NewPackageName("javax.lang.model"),
	types.NewPackageName("javax.management"),
	types.NewPackageName("javax.naming"),
	types.NewPackageName("javax.net"),
	types.NewPackageName("javax.print"),
	types.NewPackageName("javax.rmi.ssl"),
	types.NewPackageName("javax.security"),
	types.NewPackageName("javax.script"),
	types.NewPackageName("javax.smartcardio"),
	types.NewPackageName("javax.sound"),
	types.NewPackageName("javax.sql"),
	types.NewPackageName("javax.swing"),
	types.NewPackageName("javax.tools"),
	types.NewPackageName("javax.transaction.xa"),
	types.NewPackageName("javax.xml"),
	types.NewPackageName("jdk"),
	types.NewPackageName("netscape.javascript"),
	types.NewPackageName("org.ietf.jgss"),
	types.NewPackageName("org.jcp.xml.dsig.internal"),
	types.NewPackageName("org.w3c.dom"),
	types.NewPackageName("org.xml.sax"),
	types.NewPackageName("sun"),
}

func IsStdlib(imp types.PackageName) bool {
	for _, prefix := range stdlibPrefixes {
		if types.PackageNamesHasPrefix(imp, prefix) {
			return true
		}
	}
	return false
}
