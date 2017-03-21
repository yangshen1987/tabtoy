package exportorv2

import (
	"strings"

	"github.com/davyxu/tabtoy/exportorv2/i18n"
	"github.com/davyxu/tabtoy/exportorv2/model"
	"github.com/davyxu/tabtoy/util"
)

/*
	Sheet数据表单的处理

*/

type DataSheet struct {
	*Sheet
}

func (self *DataSheet) Valid() bool {

	name := strings.TrimSpace(self.Sheet.Name)
	if name != "" && name[0] == '#' {
		return false
	}

	return self.GetCellData(0, 0) != ""
}

func (self *DataSheet) Export(file *File, dataModel *model.DataModel, dataHeader, parentHeader *DataHeader) bool {

	verticalHeader := file.LocalFD.Pragma.GetBool("Vertical")

	if verticalHeader {
		return self.exportColumnMajor(file, dataModel, dataHeader, parentHeader)
	} else {
		return self.exportRowMajor(file, dataModel, dataHeader, parentHeader)
	}

}

// 导出以行数据延展的表格(普通表格)
func (self *DataSheet) exportRowMajor(file *File, dataModel *model.DataModel, dataHeader, parentHeader *DataHeader) bool {

	// 是否继续读行
	var readingLine bool = true

	var meetEmptyLine bool

	var warningAfterEmptyLineDataOnce bool

	// 遍历每一行
	for self.Row = DataSheetHeader_DataBegin; readingLine; self.Row++ {

		// 整行都是空的
		if self.IsFullRowEmpty(self.Row, dataHeader.RawFieldCount()) {

			// 再次碰空行, 表示确实是空的
			if meetEmptyLine {
				break

			} else {
				meetEmptyLine = true
			}

			continue

		} else {

			//已经碰过空行, 这里又碰到数据, 说明有人为隔出的空行, 做warning提醒, 防止数据没导出
			if meetEmptyLine && !warningAfterEmptyLineDataOnce {
				r, _ := self.GetRC()

				log.Warnf("%s %s|%s(%s)", i18n.String(i18n.DataSheet_RowDataSplitedByEmptyLine), self.file.FileName, self.Name, util.ConvR1C1toA1(r, 1))

				warningAfterEmptyLineDataOnce = true
			}

			// 曾经有过空行, 即便现在不是空行也没用, 结束
			if meetEmptyLine {
				break
			}

		}

		line := model.NewLineData()

		// 遍历每一列
		for self.Column = 0; self.Column < dataHeader.RawFieldCount(); self.Column++ {

			fieldDef := fieldDefGetter(self.Column, dataHeader, parentHeader)

			// 数据大于列头时, 结束这个列
			if fieldDef == nil {
				break
			}

			// #开头表示注释, 跳过
			if strings.Index(fieldDef.Name, "#") == 0 {
				continue
			}

			rawValue := self.GetCellData(self.Row, self.Column)

			r, c := self.GetRC()

			line.Add(&model.FieldValue{
				FieldDef:  fieldDef,
				RawValue:  rawValue,
				SheetName: self.Name,
				FileName:  self.file.FileName,
				R:         r,
				C:         c,
			})

		}

		dataModel.Add(line)

	}

	return true
}

// 多表合并时, 要从从表的字段名在主表的表头里做索引
func fieldDefGetter(index int, dataHeader, parentHeader *DataHeader) *model.FieldDescriptor {

	fieldDef := dataHeader.RawField(index)
	if fieldDef == nil {
		return nil
	}

	if parentHeader != nil {
		ret, ok := parentHeader.HeaderByName[fieldDef.Name]
		if !ok {
			return nil
		}
		return ret
	}

	return fieldDef

}

func mustFillCheck(fd *model.FieldDescriptor, raw string) bool {
	// 值重复检查
	if fd.Meta.GetBool("MustFill") {

		if raw == "" {
			log.Errorf("%s, %s", i18n.String(i18n.DataSheet_MustFill), fd.String())
			return false
		}
	}

	return true
}

func newDataSheet(sheet *Sheet) *DataSheet {

	return &DataSheet{
		Sheet: sheet,
	}
}
