package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/model"

	"github.com/gin-gonic/gin/binding"
	"github.com/xuri/excelize/v2"
)

type Template struct {
	SheetName string
	TableName string
	Columns   []Column
}

type Column struct {
	Header string
	Width  float64
	Align  string
}

var (
	userTemplate = Template{
		SheetName: "Users",
		TableName: "TableUsers",
		Columns: []Column{
			{Header: "Employee ID", Width: 20, Align: "center"},
			{Header: "Name", Width: 30, Align: "left"},
			{Header: "Role", Width: 20, Align: "center"},
		},
	}

	criticismTemplate = Template{
		SheetName: "Criticism & Feedback",
		TableName: "TableCriticism",
		Columns: []Column{
			{Header: "Subject", Width: 50, Align: "left"},
			{Header: "Body", Width: 150, Align: "left"},
			{Header: "Created At", Width: 20, Align: "center"},
		},
	}

	innovationTemplate = Template{
		SheetName: "Innovation",
		TableName: "TableInnovation",
		Columns: []Column{
			{Header: "Employee ID", Width: 20, Align: "center"},
			{Header: "Name", Width: 30, Align: "left"},
			{Header: "Subject", Width: 50, Align: "left"},
			{Header: "Body", Width: 150, Align: "left"},
			{Header: "Created At", Width: 20, Align: "center"},
		},
	}

	alignCenter = &excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	}

	alignLeft = &excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "top",
			WrapText:   true,
		},
	}
)

// Parse Users Excel
func ParseUsersExcel(excel io.Reader) ([]dto.UserRequest, error) {
	rows, err := readExcel(excel)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("Excel file is empty")
	}
	if !matchesHeader(rows[0], userTemplate.Columns) {
		return nil, errors.New("Excel columns do not match expected format")
	}

	var users []dto.UserRequest
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if isEmptyRow(row) {
			continue
		}

		var request dto.UserRequest
		var strID string

		if len(row) > 0 {
			strID = strings.TrimSpace(row[0])
		}
		if len(row) > 1 {
			request.Name = strings.TrimSpace(row[1])
		}
		if len(row) > 2 {
			request.Role = model.RoleEnum(strings.TrimSpace(row[2]))
		}

		if strID != "" {
			employeeID, err := StringToUint(strID)
			if err != nil {
				return nil, fmt.Errorf("row %d: invalid Employee ID", i+1)
			}

			request.EmployeeID = employeeID
		}
		if err := binding.Validator.ValidateStruct(request); err != nil {
			return nil, fmt.Errorf("row %d: invalid user data", i+1)
		}

		users = append(users, request)
	}

	return users, nil
}

func matchesHeader(row []string, columns []Column) bool {
	if len(row) != len(columns) {
		return false
	}
	for i, column := range columns {
		if strings.TrimSpace(row[i]) != column.Header {
			return false
		}
	}
	return true
}

func isEmptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

// Generate Users Excel
func GenerateUsersExcel(users []model.User) (*bytes.Buffer, error) {
	excel := excelize.NewFile()
	defer excel.Close()
	styleCenterID, err := excel.NewStyle(alignCenter)
	if err != nil {
		return nil, errors.New("Failed to create center style")
	}

	styleLeftID, err := excel.NewStyle(alignLeft)
	if err != nil {
		return nil, errors.New("Failed to create left style")
	}

	excel.SetSheetName("Sheet1", userTemplate.SheetName)

	for i, column := range userTemplate.Columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		excel.SetCellValue(userTemplate.SheetName, cell, column.Header)
	}

	for i, user := range users {
		row := IntToString(i + 2)
		excel.SetCellValue(userTemplate.SheetName, "A"+row, user.EmployeeID)
		excel.SetCellValue(userTemplate.SheetName, "B"+row, user.Name)
		excel.SetCellValue(userTemplate.SheetName, "C"+row, user.Role)
	}

	table := makeTable(len(userTemplate.Columns), len(users)+1, userTemplate.TableName)
	err = excel.AddTable(userTemplate.SheetName, table)
	if err != nil {
		return nil, errors.New("Failed to format table on sheet")
	}

	for i, col := range userTemplate.Columns {
		colLetter, _ := excelize.ColumnNumberToName(i + 1)
		excel.SetColWidth(userTemplate.SheetName, colLetter, colLetter, col.Width)

		if col.Align == "center" {
			excel.SetColStyle(userTemplate.SheetName, colLetter+":"+colLetter, styleCenterID)
		} else {
			excel.SetColStyle(userTemplate.SheetName, colLetter+":"+colLetter, styleLeftID)
		}
	}

	var buffer bytes.Buffer
	err = excel.Write(&buffer)
	if err != nil {
		return nil, errors.New("Failed to write to buffer")
	}

	return &buffer, nil
}

// Generate Aspirations Excel by type (criticism or innovation)
func GenerateAspirationsExcel(messages []model.Aspiration) (*bytes.Buffer, error) {
	typeGroup := make(map[model.TypeEnum][]model.Aspiration)
	for _, message := range messages {
		typeGroup[message.Type] = append(typeGroup[message.Type], message)
	}

	excel := excelize.NewFile()
	defer excel.Close()
	styleCenterID, err := excel.NewStyle(alignCenter)
	if err != nil {
		return nil, errors.New("Failed to create center style")
	}

	styleLeftID, err := excel.NewStyle(alignLeft)
	if err != nil {
		return nil, errors.New("Failed to create left style")
	}

	isFirstSheet := true
	for msgType, message := range typeGroup {
		var template Template
		switch msgType {
		case model.TypeCriticism:
			template = criticismTemplate
		case model.TypeInnovation:
			template = innovationTemplate
		default:
			return nil, errors.New("Unknown message type")
		}

		if isFirstSheet {
			excel.SetSheetName("Sheet1", template.SheetName)
			isFirstSheet = false
		} else {
			excel.NewSheet(template.SheetName)
		}

		for i, column := range template.Columns {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			excel.SetCellValue(template.SheetName, cell, column.Header)
		}

		for i, item := range message {
			row := IntToString(i + 2)
			switch msgType {
			case model.TypeCriticism:
				excel.SetCellValue(template.SheetName, "A"+row, item.Subject)
				excel.SetCellValue(template.SheetName, "B"+row, item.Body)
				excel.SetCellValue(template.SheetName, "C"+row, item.CreatedAt.Format("2006-01-02 15:04:05"))
			case model.TypeInnovation:
				excel.SetCellValue(template.SheetName, "A"+row, item.Sender.EmployeeID)
				excel.SetCellValue(template.SheetName, "B"+row, item.Sender.Name)
				excel.SetCellValue(template.SheetName, "C"+row, item.Subject)
				excel.SetCellValue(template.SheetName, "D"+row, item.Body)
				excel.SetCellValue(template.SheetName, "E"+row, item.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}

		table := makeTable(len(template.Columns), len(message)+1, template.TableName)
		err := excel.AddTable(template.SheetName, table)
		if err != nil {
			return nil, errors.New("Failed to format table on sheet")
		}

		for i, col := range template.Columns {
			colLetter, _ := excelize.ColumnNumberToName(i + 1)
			excel.SetColWidth(template.SheetName, colLetter, colLetter, col.Width)

			if col.Align == "center" {
				excel.SetColStyle(template.SheetName, colLetter+":"+colLetter, styleCenterID)
			} else {
				excel.SetColStyle(template.SheetName, colLetter+":"+colLetter, styleLeftID)
			}
		}
	}

	var buffer bytes.Buffer
	err = excel.Write(&buffer)
	if err != nil {
		return nil, errors.New("Failed to write to buffer")
	}

	return &buffer, nil
}

// helper functions
// Read Excel file
func readExcel(file io.Reader) ([][]string, error) {
	excel, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer excel.Close()

	rows, err := excel.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// Make a table
func makeTable(rowLength, colLength int, tableName string) *excelize.Table {
	lastCell, _ := excelize.CoordinatesToCellName(rowLength, colLength)

	show := true
	return &excelize.Table{
		Range:             "A1:" + lastCell,
		Name:              tableName,
		StyleName:         "TableStyleMedium9",
		ShowHeaderRow:     &show,
		ShowRowStripes:    &show,
		ShowColumnStripes: false,
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
	}
}
