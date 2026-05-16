package persistence

import (
	"github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
	"github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
)

type taskModel struct {
	database.SnowflakeBaseModel
	TaskID      string `gorm:"size:64;uniqueIndex;not null"`
	TaskType    string `gorm:"size:32;not null"`
	Type        string `gorm:"size:32;not null"`
	Expression  string `gorm:"size:128;not null"`
	Description string `gorm:"size:256"`
	Group       string `gorm:"size:64;default:default"`
	Status      int    `gorm:"default:0"`
	Params      string `gorm:"type:text"`
	RetryCount  int    `gorm:"default:3"`
}

func (taskModel) TableName() string {
	return "base_tasks"
}

func (m *taskModel) toDomain() *entity.Task {
	if m == nil {
		return nil
	}
	return &entity.Task{
		ID:          m.ID,
		TaskID:      m.TaskID,
		TaskType:    m.TaskType,
		Type:        m.Type,
		Expression:  m.Expression,
		Description: m.Description,
		Group:       m.Group,
		Status:      m.Status,
		Params:      m.Params,
		RetryCount:  m.RetryCount,
	}
}

func fromDomainTask(t *entity.Task) *taskModel {
	return &taskModel{
		SnowflakeBaseModel: database.SnowflakeBaseModel{
			ID:        t.ID,
			CreatedBy: t.CreatedBy,
			UpdatedBy: t.UpdatedBy,
		},
		TaskID:      t.TaskID,
		TaskType:    t.TaskType,
		Type:        t.Type,
		Expression:  t.Expression,
		Description: t.Description,
		Group:       t.Group,
		Status:      t.Status,
		Params:      t.Params,
		RetryCount:  t.RetryCount,
	}
}

type taskRepository struct {
	db *database.PostgresDB
}

var _ repo.TaskRepository = (*taskRepository)(nil)

func NewTaskRepository(db *database.PostgresDB) repo.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task *entity.Task) error {
	dbModel := fromDomainTask(task)
	if err := r.db.Create(dbModel).Error; err != nil {
		return err
	}
	task.ID = dbModel.ID
	return nil
}

func (r *taskRepository) Update(task *entity.Task) error {
	dbModel := fromDomainTask(task)
	return r.db.Where("id = ?", task.ID).Updates(dbModel).Error
}

func (r *taskRepository) GetByID(id int64) (*entity.Task, error) {
	var dbModel taskModel
	result := r.db.Where("id = ?", id).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

func (r *taskRepository) GetByTaskID(taskID string) (*entity.Task, error) {
	var dbModel taskModel
	result := r.db.Where("task_id = ?", taskID).First(&dbModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return dbModel.toDomain(), nil
}

func (r *taskRepository) List(status *int, page, pageSize int) ([]*entity.Task, int64, error) {
	var total int64
	var dbModels []taskModel

	query := r.db.Model(&taskModel{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]*entity.Task, len(dbModels))
	for i, model := range dbModels {
		tasks[i] = model.toDomain()
	}

	return tasks, total, nil
}

func (r *taskRepository) ListByType(taskType string, status *int) ([]*entity.Task, error) {
	var dbModels []taskModel

	query := r.db.Model(&taskModel{}).Where("task_type = ?", taskType)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, err
	}

	tasks := make([]*entity.Task, len(dbModels))
	for i, model := range dbModels {
		tasks[i] = model.toDomain()
	}

	return tasks, nil
}

func (r *taskRepository) ListAll() ([]*entity.Task, error) {
	var dbModels []taskModel
	if err := r.db.Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, err
	}

	tasks := make([]*entity.Task, len(dbModels))
	for i, model := range dbModels {
		tasks[i] = model.toDomain()
	}

	return tasks, nil
}

func (r *taskRepository) UpdateStatus(taskID string, status int) error {
	return r.db.Model(&taskModel{}).Where("task_id = ?", taskID).Update("status", status).Error
}

func (r *taskRepository) Delete(taskID string) error {
	return r.db.Where("task_id = ?", taskID).Delete(&taskModel{}).Error
}

func (r *taskRepository) GetEnabledTasks() ([]*entity.Task, error) {
	var dbModels []taskModel
	if err := r.db.Where("status = ?", 1).Order("created_at DESC").Find(&dbModels).Error; err != nil {
		return nil, err
	}

	tasks := make([]*entity.Task, len(dbModels))
	for i, model := range dbModels {
		tasks[i] = model.toDomain()
	}

	return tasks, nil
}
