package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.io.Reader;
import java.io.StringReader;
import java.io.StringWriter;
import java.io.Writer;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.parsers.ParserConfigurationException;
import javax.xml.stream.XMLOutputFactory;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamWriter;
import org.junit.jupiter.api.Test;
import org.w3c.dom.Document;
import org.w3c.dom.Node;
import org.xml.sax.InputSource;
import org.xml.sax.SAXException;

public class SafeXmlTest {

  @Test
  public void properlyEscapesCDataSection()
      throws XMLStreamException, ParserConfigurationException, IOException, SAXException {
    try (Writer writer = new StringWriter()) {
      XMLStreamWriter xml = XMLOutputFactory.newDefaultFactory().createXMLStreamWriter(writer);

      xml.writeStartDocument("UTF-8", "1.0");
      // Output the "end of cdata" marker
      SafeXml.writeTextElement(xml, "container", "]]>");
      xml.writeEndDocument();

      DocumentBuilderFactory factory = DocumentBuilderFactory.newInstance();
      DocumentBuilder builder;
      try (Reader reader = new StringReader(writer.toString())) {
        builder = factory.newDocumentBuilder();
        Document parsed = builder.parse(new InputSource(reader));

        Node container = parsed.getElementsByTagName("container").item(0);

        assertEquals("]]>", container.getTextContent());
      }
    }
  }
}
