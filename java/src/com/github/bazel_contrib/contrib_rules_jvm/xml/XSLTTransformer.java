import javax.xml.transform.Transformer;
import javax.xml.transform.TransformerFactory;
import javax.xml.transform.stream.StreamResult;
import javax.xml.transform.stream.StreamSource;
import java.io.File;

public class XSLTTransformer {
    public static void main(String[] args) throws Exception {
        if (args.length != 2) {
            System.err.println("Usage: java XSLTTransformer <input.xml> <transform.xslt>");
            System.exit(1);
        }

        String inputXML = args[0];
        String xsltFile = args[1];

        // Create transformer factory
        TransformerFactory factory = TransformerFactory.newInstance();

        // Load the XSLT file
        StreamSource xslt = new StreamSource(new File(xsltFile));

        // Create a transformer
        Transformer transformer = factory.newTransformer(xslt);

        // Load the input XML file
        StreamSource xmlInput = new StreamSource(new File(inputXML));

        // Set the output file
        StreamResult xmlOutput = new StreamResult(System.out);

        // Perform the transformation
        transformer.transform(xmlInput, xmlOutput);
    }
}
