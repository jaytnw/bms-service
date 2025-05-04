package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jaytnw/bms-service/internal/apperr"
	"github.com/jaytnw/bms-service/internal/models"
	"github.com/jaytnw/bms-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type StatusService interface {
	GetAllStatus(ctx context.Context) ([]models.Status, error)
	HandleMQTTStatusUpdate(topic string, payload []byte)
	GetStatusByWasherID(ctx context.Context, washerID string) (*models.Status, error)
	GetStatusHistoryByWasherID(ctx context.Context, washerID string) ([]models.Status, error)
	GetDormStatusReport(ctx context.Context) ([]models.DormStatusReport, error)
}

type statusService struct {
	statusRepo  repository.StatusRepository
	externalAPI ExternalAPIService
	redisClient *redis.Client
}

func NewStatusService(repo repository.StatusRepository, api ExternalAPIService, redisClient *redis.Client) StatusService {
	return &statusService{
		statusRepo:  repo,
		externalAPI: api,
		redisClient: redisClient,
	}
}

func (s *statusService) GetAllStatus(ctx context.Context) ([]models.Status, error) {
	statuses, err := s.statusRepo.FindAll(ctx)
	if err != nil {
		log.Printf("[StatusService] failed to FindAll: %v", err)
	}

	return statuses, err
}

func (s *statusService) HandleMQTTStatusUpdate(topic string, payload []byte) {
	parts := strings.Split(topic, "/")
	if len(parts) < 4 {
		log.Printf("❌ Invalid topic format: %s", topic)
		return
	}

	dormId := parts[1]
	washerId := parts[2]

	status := &models.Status{
		DormID:   dormId,
		WasherID: washerId,
		Status:   string(payload),
	}

	if err := s.statusRepo.SaveStatus(context.Background(), status); err != nil {
		log.Printf("❌ Failed to save status: %v", err)
	}
}

func (s *statusService) GetStatusByWasherID(ctx context.Context, washerID string) (*models.Status, error) {
	status, err := s.statusRepo.FindLatestByWasherID(ctx, washerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New("NOT_FOUND", "Status not found", 404, err)
		}
		return nil, apperr.New("DB_ERROR", "Failed to get status", 500, err)
	}
	return status, nil
}

func (s *statusService) GetStatusHistoryByWasherID(ctx context.Context, washerID string) ([]models.Status, error) {
	statuses, err := s.statusRepo.FindHistoryByWasherID(ctx, washerID)
	if err != nil {
		return nil, apperr.New("DB_ERROR", "Failed to get status history", 500, err)
	}

	if len(statuses) == 0 {
		return nil, apperr.New("NOT_FOUND", "No status history found", 404, nil)
	}

	return statuses, nil
}

// func (s *statusService) GetDormStatusReport(ctx context.Context) ([]models.DormStatusReport, error) {
// 	start := time.Now()

// 	// 📡 ดึงจาก API ภายนอก
// 	machines, err := s.externalAPI.FetchWashingMachines()
// 	if err != nil {
// 		return nil, apperr.New("API_ERROR", "Failed to fetch machines", 500, err)
// 	}
// 	log.Printf("📡 FetchWashingMachines took: %v", time.Since(start))

// 	// 🧼 ดึง history ทีเดียว
// 	washerIDs := extractWasherIDs(machines)
// 	historyStart := time.Now()
// 	allHistories, err := s.statusRepo.FindHistoryByWasherIDs(ctx, washerIDs)
// 	if err != nil {
// 		return nil, apperr.New("DB_ERROR", "Failed to fetch histories", 500, err)
// 	}

// 	// 🧠 จัดกลุ่ม washerID → []Status
// 	historyMap := make(map[string][]models.Status)
// 	for _, h := range allHistories {
// 		historyMap[h.WasherID] = append(historyMap[h.WasherID], h)
// 	}
// 	log.Printf("📦 Grouping + History took: %v", time.Since(historyStart))

// 	// 🏠 จัดกลุ่ม dorm
// 	grouped := make(map[string]*models.DormStatusReport)
// 	for _, m := range machines {
// 		history := historyMap[m.IDWashingMachine]
// 		if len(history) == 0 {
// 			continue
// 		}

// 		if _, ok := grouped[m.IDDorm]; !ok {
// 			grouped[m.IDDorm] = &models.DormStatusReport{
// 				DormID:   m.IDDorm,
// 				DormName: m.DormName,
// 				Machines: []models.WasherStatusHistory{},
// 			}
// 		}

// 		grouped[m.IDDorm].Machines = append(grouped[m.IDDorm].Machines, models.WasherStatusHistory{
// 			WasherID: m.IDWashingMachine,
// 			History:  history,
// 		})
// 	}

// 	// 🧾 แปลง map เป็น slice
// 	convertStart := time.Now()
// 	result := make([]models.DormStatusReport, 0, len(grouped))
// 	for _, report := range grouped {
// 		result = append(result, *report)
// 	}
// 	log.Printf("📊 Convert map to slice took: %v", time.Since(convertStart))
// 	log.Printf("✅ Total GetDormStatusReport took: %v", time.Since(start))

// 	return result, nil
// }

func (s *statusService) GetDormStatusReport(ctx context.Context) ([]models.DormStatusReport, error) {

	totalStart := time.Now()
	const cacheDuration = 24 * time.Hour

	// ⏱️ Step 1: ลองดึง machines จาก Redis
	var machines []WashingMachine
	const cacheKey = "external_api:washing_machines"
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cached), &machines); err == nil {
			log.Println("⚡ Loaded WashingMachines from Redis")
		}
	}

	// 🔁 ถ้า Redis ไม่มี หรือ error ให้เรียก API แล้ว cache ใหม่
	if len(machines) == 0 {
		machines, err = s.externalAPI.FetchWashingMachines()
		if err != nil {
			return nil, apperr.New("API_ERROR", "Failed to fetch machines", 500, err)
		}

		bytes, _ := json.Marshal(machines)

		s.redisClient.Set(ctx, cacheKey, bytes, cacheDuration)
		log.Println("✅ Cached WashingMachines to Redis")
	}

	log.Printf("📡 FetchWashingMachines took: %v", time.Since(totalStart))

	// ⏱️ Step 2: Load history
	start := time.Now()
	washerIDs := make([]string, 0, len(machines))
	machineMap := make(map[string]WashingMachine)
	for _, m := range machines {
		washerIDs = append(washerIDs, m.IDWashingMachine)
		machineMap[m.IDWashingMachine] = m
	}

	histories, err := s.statusRepo.FindLatest50HistoryByWasherIDs(ctx, washerIDs)
	if err != nil {
		return nil, apperr.New("DB_ERROR", "Failed to fetch histories", 500, err)
	}

	// ✏️ Step 3: Group & Merge
	grouped := make(map[string]*models.DormStatusReport)
	for _, h := range histories {
		m := machineMap[h.WasherID]
		report, ok := grouped[m.IDDorm]
		if !ok {
			report = &models.DormStatusReport{
				DormID:   m.IDDorm,
				DormName: m.DormName,
				Machines: []models.WasherStatusHistory{},
			}
			grouped[m.IDDorm] = report
		}

		found := false
		for i := range report.Machines {
			if report.Machines[i].WasherID == h.WasherID {
				report.Machines[i].History = append(report.Machines[i].History, models.ToStatusDTO(h))
				found = true
				break
			}
		}
		if !found {
			report.Machines = append(report.Machines, models.WasherStatusHistory{
				WasherID: h.WasherID,
				History:  []models.StatusDTO{models.ToStatusDTO(h)},
			})
		}
	}
	log.Printf("📦 Grouping + History took: %v", time.Since(start))

	start = time.Now()
	result := make([]models.DormStatusReport, 0, len(machines))
	added := make(map[string]bool)
	for _, m := range machines {
		if report, ok := grouped[m.IDDorm]; ok {
			if !added[m.IDDorm] {
				result = append(result, *report)
				added[m.IDDorm] = true
			}
		}
	}
	log.Printf("📊 Convert map to slice took: %v", time.Since(start))
	log.Printf("✅ Total GetDormStatusReport took: %v", time.Since(totalStart))

	return result, nil
}

// func extractWasherIDs(machines []WashingMachine) []string {
// 	ids := make([]string, 0, len(machines))
// 	for _, m := range machines {
// 		ids = append(ids, m.IDWashingMachine)
// 	}
// 	return ids
// }

// func (s *statusService) DebugExternalAPI() {
// 	data, err := s.externalAPI.FetchWashingMachines()
// 	if err != nil {
// 		log.Printf("❌ Error fetching washing machines: %v", err)
// 		return
// 	}

// 	// สร้าง map เพื่อจัดกลุ่มตามหอพัก
// 	grouped := make(map[string][]string) // map[dormName][]washingMachineID

// 	for _, wm := range data {
// 		grouped[wm.DormName] = append(grouped[wm.DormName], wm.IDWashingMachine)
// 	}

// 	// แสดงผล
// 	for dormName, machineIDs := range grouped {
// 		log.Printf("🏢 Dorm: %s", dormName)
// 		for _, id := range machineIDs {
// 			log.Printf("   🧺 Machine ID: %s", id)
// 		}
// 	}
// }
