package pptx

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// Slide dimensions in EMU (English Metric Units)
// 1 inch = 914400 EMU, standard slide = 10" x 7.5"
const (
	slideWidthEMU  = 12192000 // 10 inches
	slideHeightEMU = 6858000  // 7.5 inches
)

// Builder creates PPTX files from slide images
type Builder struct{}

// NewBuilder creates a new PPTX builder
func NewBuilder() *Builder {
	return &Builder{}
}

// ExportPPTX assembles slide images into a PPTX file
func (b *Builder) ExportPPTX(slideImages []string, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// Write static structure files
	if err := writeContentTypes(w, len(slideImages)); err != nil {
		return err
	}
	if err := writeRootRels(w); err != nil {
		return err
	}
	if err := writePresentationXML(w, len(slideImages)); err != nil {
		return err
	}
	if err := writePresentationRels(w, len(slideImages)); err != nil {
		return err
	}
	if err := writeSlideMaster(w); err != nil {
		return err
	}
	if err := writeSlideMasterRels(w); err != nil {
		return err
	}
	if err := writeSlideLayout(w); err != nil {
		return err
	}
	if err := writeSlideLayoutRels(w); err != nil {
		return err
	}

	// Write each slide and its image
	for i, imgPath := range slideImages {
		slideNum := i + 1

		if err := writeSlideXML(w, slideNum); err != nil {
			return err
		}
		if err := writeSlideRels(w, slideNum); err != nil {
			return err
		}
		if err := writeImageFile(w, slideNum, imgPath); err != nil {
			return err
		}
	}

	return nil
}

// ExportPDF assembles slide images into a PDF file
func (b *Builder) ExportPDF(slideImages []string, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return buildPDF(slideImages, outputPath)
}

func addFileToZip(w *zip.Writer, name string, content string) error {
	fw, err := w.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create %s in zip: %w", name, err)
	}
	_, err = fw.Write([]byte(content))
	return err
}

func writeContentTypes(w *zip.Writer, slideCount int) error {
	overrides := ""
	for i := 1; i <= slideCount; i++ {
		overrides += fmt.Sprintf(`  <Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
`, i)
	}
	for i := 1; i <= slideCount; i++ {
		overrides += fmt.Sprintf(`  <Override PartName="/ppt/media/image%d.png" ContentType="image/png"/>
`, i)
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="png" ContentType="image/png"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
  <Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>
  <Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>
` + overrides + `</Types>`

	return addFileToZip(w, "[Content_Types].xml", content)
}

func writeRootRels(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
</Relationships>`

	return addFileToZip(w, "_rels/.rels", content)
}

func writePresentationXML(w *zip.Writer, slideCount int) error {
	slideList := ""
	for i := 1; i <= slideCount; i++ {
		slideList += fmt.Sprintf(`    <p:sldId id="%d" r:id="rId%d"/>
`, 255+i, i+2) // rId1=slideMaster, rId2=slideLayout, rId3+=slides
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:sldMasterIdLst>
    <p:sldMasterId id="2147483648" r:id="rId1"/>
  </p:sldMasterIdLst>
  <p:sldIdLst>
` + slideList + `  </p:sldIdLst>
  <p:sldSz cx="` + fmt.Sprintf("%d", slideWidthEMU) + `" cy="` + fmt.Sprintf("%d", slideHeightEMU) + `"/>
  <p:notesSz cx="` + fmt.Sprintf("%d", slideHeightEMU) + `" cy="` + fmt.Sprintf("%d", slideWidthEMU) + `"/>
</p:presentation>`

	return addFileToZip(w, "ppt/presentation.xml", content)
}

func writePresentationRels(w *zip.Writer, slideCount int) error {
	rels := `  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>
`
	for i := 1; i <= slideCount; i++ {
		rels += fmt.Sprintf(`  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>
`, i+2, i)
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
` + rels + `</Relationships>`

	return addFileToZip(w, "ppt/_rels/presentation.xml.rels", content)
}

func writeSlideMaster(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:bg>
      <p:bgPr>
        <a:solidFill><a:srgbClr val="FFFFFF"/></a:solidFill>
        <a:effectLst/>
      </p:bgPr>
    </p:bg>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
  <p:sldLayoutIdLst>
    <p:sldLayoutId id="2147483649" r:id="rId1"/>
  </p:sldLayoutIdLst>
</p:sldMaster>`

	return addFileToZip(w, "ppt/slideMasters/slideMaster1.xml", content)
}

func writeSlideMasterRels(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
</Relationships>`

	return addFileToZip(w, "ppt/slideMasters/_rels/slideMaster1.xml.rels", content)
}

func writeSlideLayout(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
  type="blank">
  <p:cSld name="Blank">
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
</p:sldLayout>`

	return addFileToZip(w, "ppt/slideLayouts/slideLayout1.xml", content)
}

func writeSlideLayoutRels(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>
</Relationships>`

	return addFileToZip(w, "ppt/slideLayouts/_rels/slideLayout1.xml.rels", content)
}

func writeSlideXML(w *zip.Writer, slideNum int) error {
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
  xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
  xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
      <p:pic>
        <p:nvPicPr>
          <p:cNvPr id="2" name="Slide Image %d"/>
          <p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr>
          <p:nvPr/>
        </p:nvPicPr>
        <p:blipFill>
          <a:blip r:embed="rId2"/>
          <a:stretch><a:fillRect/></a:stretch>
        </p:blipFill>
        <p:spPr>
          <a:xfrm>
            <a:off x="0" y="0"/>
            <a:ext cx="%d" cy="%d"/>
          </a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
        </p:spPr>
      </p:pic>
    </p:spTree>
  </p:cSld>
</p:sld>`, slideNum, slideWidthEMU, slideHeightEMU)

	return addFileToZip(w, fmt.Sprintf("ppt/slides/slide%d.xml", slideNum), content)
}

func writeSlideRels(w *zip.Writer, slideNum int) error {
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image%d.png"/>
</Relationships>`, slideNum)

	return addFileToZip(w, fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slideNum), content)
}

func writeImageFile(w *zip.Writer, slideNum int, imgPath string) error {
	imgData, err := os.ReadFile(imgPath)
	if err != nil {
		return fmt.Errorf("failed to read image %s: %w", imgPath, err)
	}

	fw, err := w.Create(fmt.Sprintf("ppt/media/image%d.png", slideNum))
	if err != nil {
		return fmt.Errorf("failed to create image in zip: %w", err)
	}

	_, err = fw.Write(imgData)
	return err
}
