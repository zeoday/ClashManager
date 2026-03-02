package repository

import (
	"fmt"

	"clash-manager/internal/model"
)

type RuleListParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Type     string `json:"type"`
	Keyword  string `json:"keyword"`
	Target   string `json:"target"` // 过滤目标名称
}

type RuleListResult struct {
	Rules      []model.Rule `json:"rules"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	TotalPages int          `json:"totalPages"`
}

type RuleRepository struct{}

func (r *RuleRepository) Create(rule *model.Rule) error {
	return DB.Create(rule).Error
}

func (r *RuleRepository) BatchCreate(rules *[]model.Rule) error {
	if len(*rules) == 0 {
		return nil
	}
	return DB.Create(&rules).Error
}

func (r *RuleRepository) FindAll() ([]model.Rule, error) {
	var rules []model.Rule
	// Order by precedence if we add a Priority field, otherwise ID
	err := DB.Order("priority asc, id asc").Find(&rules).Error
	return rules, err
}

// FindWithPagination returns rules with pagination and filtering
func (r *RuleRepository) FindWithPagination(params *RuleListParams) (*RuleListResult, error) {
	var rules []model.Rule
	var total int64

	query := DB.Model(&model.Rule{})

	fmt.Printf("[FindWithPagination] Input params - Page: %d, PageSize: %d, Type: %s, Keyword: %s, Target: %s\n",
		params.Page, params.PageSize, params.Type, params.Keyword, params.Target)

	// Filter by type
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
		fmt.Printf("[FindWithPagination] Applied type filter: %s\n", params.Type)
	}

	// Filter by keyword (search in Payload, Target, Remark)
	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		query = query.Where("payload LIKE ? OR target LIKE ? OR remark LIKE ?", keyword, keyword, keyword)
		fmt.Printf("[FindWithPagination] Applied keyword filter: %s\n", params.Keyword)
	}

	// Filter by target name
	if params.Target != "" {
		query = query.Where("target = ?", params.Target)
		fmt.Printf("[FindWithPagination] Applied target filter: %s\n", params.Target)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		fmt.Printf("[FindWithPagination] Count error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[FindWithPagination] Total count: %d\n", total)

	// Calculate offset
	offset := (params.Page - 1) * params.PageSize

	// Fetch data
	if err := query.Order("priority asc, id asc").Offset(offset).Limit(params.PageSize).Find(&rules).Error; err != nil {
		fmt.Printf("[FindWithPagination] Find error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[FindWithPagination] Fetched %d rules\n", len(rules))

	// Calculate total pages
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return &RuleListResult{
		Rules:      rules,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *RuleRepository) Delete(id uint) error {
	return DB.Delete(&model.Rule{}, id).Error
}

func (r *RuleRepository) Update(rule *model.Rule) error {
	return DB.Save(rule).Error
}

func (r *RuleRepository) FindByID(id uint) (*model.Rule, error) {
	var rule model.Rule
	err := DB.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}
