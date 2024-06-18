<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output encoding="UTF-8" method="xml"></xsl:output>

  <xsl:template match="/">
    <testsuite>
      <xsl:attribute name="tests">
        <xsl:value-of select="count(.//file)" />
      </xsl:attribute>
      <xsl:attribute name="failures">
        <xsl:value-of select="count(.//error[@severity='error'])" />
      </xsl:attribute>
      <xsl:for-each select="//checkstyle">
        <xsl:apply-templates />
      </xsl:for-each>
    </testsuite>
  </xsl:template>

  <xsl:template match="file">
    <testcase>
      <xsl:attribute name="classname">
        <xsl:value-of select="@name" />
      </xsl:attribute>
      <xsl:attribute name="name">
        <xsl:value-of select="@name" />
      </xsl:attribute>
      <xsl:apply-templates select="node()" />
    </testcase>
  </xsl:template>

  <xsl:template match="error">
    <error>
      <xsl:attribute name="type">
        <xsl:value-of select="@source" />
      </xsl:attribute>
      <xsl:attribute name="message">
        <xsl:text>Line </xsl:text>
        <xsl:value-of select="@line" />
        <xsl:text>: </xsl:text>
        <xsl:value-of select="@message" />
      </xsl:attribute>
    </error>
  </xsl:template>
</xsl:stylesheet>
