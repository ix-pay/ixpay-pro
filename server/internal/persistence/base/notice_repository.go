package persistence

import (
	"time"

	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
	"github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// noticeModel 公告数据库模型
type noticeModel struct {
	database.SnowflakeBaseModel
	Title       string     `gorm:"size:200;not null"`
	Content     string     `gorm:"type:text;not null"`
	Type        int        `gorm:"default:1"`
	Status      int        `gorm:"default:0"`
	PublisherID int64      `gorm:"index"`
	PublishTime *time.Time `gorm:"index"`
	ViewCount   int64      `gorm:"default:0"`
	IsTop       bool       `gorm:"default:false"`
	Sort        int        `gorm:"default:0"`
	Description string     `gorm:"size:500"`
}

// TableName 指定表名
func (noticeModel) TableName() string {
	return "base_notices"
}

// toDomain 将数据库模型转换为领域实体
func (m *noticeModel) toDomain() *entity.Notice {
	if m == nil {
		return nil
	}
	return &entity.Notice{
		ID:          common.ToString(m.ID),
		Title:       m.Title,
		Content:     m.Content,
		Type:        entity.NoticeType(m.Type),
		Status:      entity.NoticeStatus(m.Status),
		PublisherID: common.ToString(m.PublisherID),
		PublishTime: m.PublishTime,
		ViewCount:   m.ViewCount,
		IsTop:       m.IsTop,
		Sort:        m.Sort,
		Description: m.Description,
		CreatedBy:   common.ToString(m.CreatedBy),
		CreatedAt:   m.CreatedAt,
		UpdatedBy:   common.ToString(m.UpdatedBy),
		UpdatedAt:   m.UpdatedAt,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainNotice(notice *entity.Notice) (*noticeModel, error) {
	id, createdBy, updatedBy := common.SetBaseFields(notice.ID, notice.CreatedBy, notice.UpdatedBy)

	return &noticeModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        id,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
		Title:       notice.Title,
		Content:     notice.Content,
		Type:        int(notice.Type),
		Status:      int(notice.Status),
		PublisherID: common.TryParseInt64(notice.PublisherID),
		PublishTime: notice.PublishTime,
		ViewCount:   notice.ViewCount,
		IsTop:       notice.IsTop,
		Sort:        notice.Sort,
		Description: notice.Description,
	}, nil
}

// noticeReadRecordModel 公告阅读记录数据库模型
type noticeReadRecordModel struct {
	database.SnowflakeBaseModel
	NoticeID int64     `gorm:"index;not null"`
	UserID   int64     `gorm:"index;not null"`
	ReadTime time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (noticeReadRecordModel) TableName() string {
	return "base_notice_read_records"
}

// toDomain 将数据库模型转换为领域实体
func (m *noticeReadRecordModel) toDomain() *entity.NoticeReadRecord {
	if m == nil {
		return nil
	}
	return &entity.NoticeReadRecord{
		ID:       common.ToString(m.ID),
		NoticeID: common.ToString(m.NoticeID),
		UserID:   common.ToString(m.UserID),
		ReadTime: m.ReadTime,
	}
}

// fromDomain 将领域实体转换为数据库模型
func fromDomainNoticeReadRecord(record *entity.NoticeReadRecord) (*noticeReadRecordModel, error) {
	id, err := common.ParseInt64(record.ID)
	if err != nil {
		return nil, err
	}

	noticeID, _ := common.ParseInt64(record.NoticeID)
	userID, _ := common.ParseInt64(record.UserID)

	return &noticeReadRecordModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID: id,
		},
		NoticeID: noticeID,
		UserID:   userID,
		ReadTime: record.ReadTime,
	}, nil
}

// noticeRepository Repository 实现
type noticeRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.NoticeRepository = (*noticeRepository)(nil)

// NewNoticeRepository 创建公告仓库实现
func NewNoticeRepository(db *database.PostgresDB) repo.NoticeRepository {
	return &noticeRepository{db: db}
}

// GetByID 根据 ID 查询公告
func (r *noticeRepository) GetByID(id string) (*entity.Notice, error) {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return nil, err
	}

	var dbModel noticeModel
	result := r.db.Where("id = ?", intID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// Create 创建公告
func (r *noticeRepository) Create(notice *entity.Notice) error {
	dbModel, err := fromDomainNotice(notice)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
}

// Update 更新公告
func (r *noticeRepository) Update(notice *entity.Notice) error {
	dbModel, err := fromDomainNotice(notice)
	if err != nil {
		return err
	}

	return r.db.Save(dbModel).Error
}

// Delete 删除公告
func (r *noticeRepository) Delete(id string) error {
	intID, err := common.ParseInt64(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&noticeModel{}, intID).Error
}

// List 分页查询公告列表
func (r *noticeRepository) List(page, pageSize int, filters map[string]interface{}) ([]*entity.Notice, int64, error) {
	var total int64
	var dbModels []noticeModel

	query := r.db.Model(&noticeModel{})

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("is_top DESC, sort ASC, publish_time DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	notices := make([]*entity.Notice, len(dbModels))
	for i, model := range dbModels {
		notices[i] = model.toDomain()
	}

	return notices, total, nil
}

// GetPublishedList 获取已发布公告列表
func (r *noticeRepository) GetPublishedList(page, pageSize int, filters map[string]interface{}) ([]*entity.Notice, int64, error) {
	var total int64
	var dbModels []noticeModel

	query := r.db.Model(&noticeModel{}).Where("status = ?", 1)

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("is_top DESC, sort ASC, publish_time DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	notices := make([]*entity.Notice, len(dbModels))
	for i, model := range dbModels {
		notices[i] = model.toDomain()
	}

	return notices, total, nil
}

// IncrementViewCount 增加浏览次数
func (r *noticeRepository) IncrementViewCount(id string) error {
	intID := common.TryParseInt64(id)
	return r.db.Exec("UPDATE base_notices SET view_count = view_count + 1 WHERE id = ?", intID).Error
}

// GetStatistics 获取公告统计信息
func (r *noticeRepository) GetStatistics() (*entity.NoticeStatistics, error) {
	var total, published, draft, archived int64

	// 总数
	if err := r.db.Model(&noticeModel{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 已发布
	if err := r.db.Model(&noticeModel{}).Where("status = ?", 1).Count(&published).Error; err != nil {
		return nil, err
	}

	// 草稿
	if err := r.db.Model(&noticeModel{}).Where("status = ?", 0).Count(&draft).Error; err != nil {
		return nil, err
	}

	// 已归档
	if err := r.db.Model(&noticeModel{}).Where("status = ?", 2).Count(&archived).Error; err != nil {
		return nil, err
	}

	return &entity.NoticeStatistics{
		TotalCount:     total,
		PublishedCount: published,
		DraftCount:     draft,
		ArchivedCount:  archived,
	}, nil
}

// noticeReadRecordRepository Repository 实现
type noticeReadRecordRepository struct {
	db *database.PostgresDB
}

// 确保实现接口
var _ repo.NoticeReadRecordRepository = (*noticeReadRecordRepository)(nil)

// NewNoticeReadRecordRepository 创建公告阅读记录仓库实现
func NewNoticeReadRecordRepository(db *database.PostgresDB) repo.NoticeReadRecordRepository {
	return &noticeReadRecordRepository{db: db}
}

// Create 创建阅读记录
func (r *noticeReadRecordRepository) Create(record *entity.NoticeReadRecord) error {
	dbModel, err := fromDomainNoticeReadRecord(record)
	if err != nil {
		return err
	}

	return r.db.Create(dbModel).Error
}

// CreateOrUpdate 创建或更新阅读记录
func (r *noticeReadRecordRepository) CreateOrUpdate(noticeID, userID string) error {
	// 先检查是否已存在
	existing, _ := r.GetByNoticeIDAndUserID(noticeID, userID)
	if existing != nil {
		return nil // 已存在则不重复创建
	}

	record := &entity.NoticeReadRecord{
		NoticeID: noticeID,
		UserID:   userID,
		ReadTime: time.Now(),
	}

	return r.Create(record)
}

// GetByNoticeIDAndUserID 根据公告 ID 和用户 ID 查询阅读记录
func (r *noticeReadRecordRepository) GetByNoticeIDAndUserID(noticeID, userID string) (*entity.NoticeReadRecord, error) {
	intNoticeID := common.TryParseInt64(noticeID)
	intUserID := common.TryParseInt64(noticeID)

	var dbModel noticeReadRecordModel
	result := r.db.Where("notice_id = ? AND user_id = ?", intNoticeID, intUserID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}

	return dbModel.toDomain(), nil
}

// GetReadUserCount 获取公告的阅读用户数
func (r *noticeReadRecordRepository) GetReadUserCount(noticeID string) (int64, error) {
	intNoticeID := common.TryParseInt64(noticeID)
	var count int64
	result := r.db.Model(&noticeReadRecordModel{}).Where("notice_id = ?", intNoticeID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}
