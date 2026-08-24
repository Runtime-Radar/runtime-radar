package template

func Init(templatesFolder string) {
	initMitreMatrixData(templatesFolder)
	initHTMLs(templatesFolder)
	initTexts(templatesFolder)
}
