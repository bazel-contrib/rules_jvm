<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output encoding="UTF-8" method="xml"></xsl:output>

  <xsl:template match="/">
    <testsuites>
      <testsuite package="Checkstyle">
        <xsl:attribute name="name">
          <xsl:value-of select="//checkstyle/file/@name" />
        </xsl:attribute>
        <xsl:attribute name="tests">
          <xsl:value-of select="count(.//error)" />
        </xsl:attribute>
        <xsl:attribute name="errors">
          <xsl:value-of select="count(.//error)" />
        </xsl:attribute>
        <xsl:for-each select="//checkstyle">
          <xsl:apply-templates />
        </xsl:for-each>
      </testsuite>
    </testsuites>
  </xsl:template>

  <xsl:template match="error">
    <testcase>
      <xsl:attribute name="name">
        <xsl:value-of select="@source" />
      </xsl:attribute>
      <xsl:attribute name="classname">
        <xsl:value-of select="../@name" />
      </xsl:attribute>
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
    </testcase>
  </xsl:template>
  
</xsl:stylesheet>
