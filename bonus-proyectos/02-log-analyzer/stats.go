package main

import (
	"sort"
	"sync"
	"sync/atomic"
)

// AttackCount representa el conteo de un tipo de ataque específico.
type AttackCount struct {
	Name  string
	Count int64
}

// IPCount representa el conteo de ataques desde una IP específica.
type IPCount struct {
	IP    string
	Count int64
}

// StatsSnapshot es una copia inmutable del estado actual de las estadísticas.
// Se usa para generar reportes sin bloquear el mapa principal.
type StatsSnapshot struct {
	TotalLines   int64
	TotalAttacks int64
	UniqueIPs    int
	BySeverity   [4]int64      // Índices: 0=LOW, 1=MEDIUM, 2=HIGH, 3=CRITICAL
	TopAttacks   []AttackCount // Ataques ordenados por frecuencia descendente
	TopIPs       []IPCount     // IPs ordenadas por frecuencia descendente
}

// Stats es la estructura thread-safe que almacena todas las estadísticas
// de detección. Usa sync.RWMutex para permitir lecturas concurrentes
// seguras mientras los workers escriben los contadores.
type Stats struct {
	mu sync.RWMutex

	// Contadores protegidos por el mutex
	attackCounts map[string]int64 // Patrón → cantidad de detecciones
	ipCounts     map[string]int64 // IP → cantidad de ataques
	bySeverity   [4]int64         // Conteo por nivel de severidad

	// Contadores atómicos (no requieren el mutex para lectura/escritura)
	totalLines   atomic.Int64
	totalAttacks atomic.Int64
}

// NewStats crea una nueva instancia de estadísticas inicializada.
func NewStats() *Stats {
	return &Stats{
		attackCounts: make(map[string]int64),
		ipCounts:     make(map[string]int64),
	}
}

// IncrementLines incrementa el contador total de líneas procesadas.
// Usa atomic para máxima concurrencia sin bloqueos.
func (s *Stats) IncrementLines() {
	s.totalLines.Add(1)
}

// IncrementLinesBy incrementa el contador total de líneas por un valor dado.
func (s *Stats) IncrementLinesBy(n int64) {
	s.totalLines.Add(n)
}

// RecordAttack registra un ataque detectado. Actualiza los contadores
// de patrón, IP y severidad de forma thread-safe usando RWMutex.
func (s *Stats) RecordAttack(pattern string, severity Severity, ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalAttacks.Add(1)

	// Actualizar contador del patrón
	s.attackCounts[pattern]++

	// Actualizar contador de severidad
	if severity >= 0 && int(severity) < len(s.bySeverity) {
		s.bySeverity[severity]++
	}

	// Actualizar contador de IP
	if ip != "" {
		s.ipCounts[ip]++
	}
}

// Snapshot genera una copia inmutable del estado actual de las estadísticas.
// Usa un read lock para no bloquear a los workers que están escribiendo.
func (s *Stats) Snapshot() StatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := StatsSnapshot{
		TotalLines:   s.totalLines.Load(),
		TotalAttacks: s.totalAttacks.Load(),
		UniqueIPs:    len(s.ipCounts),
		BySeverity:   s.bySeverity,
	}

	// Construir lista ordenada de ataques por frecuencia
	snapshot.TopAttacks = make([]AttackCount, 0, len(s.attackCounts))
	for name, count := range s.attackCounts {
		snapshot.TopAttacks = append(snapshot.TopAttacks, AttackCount{
			Name:  name,
			Count: count,
		})
	}
	sort.Slice(snapshot.TopAttacks, func(i, j int) bool {
		return snapshot.TopAttacks[i].Count > snapshot.TopAttacks[j].Count
	})

	// Construir lista ordenada de IPs por frecuencia
	snapshot.TopIPs = make([]IPCount, 0, len(s.ipCounts))
	for ip, count := range s.ipCounts {
		snapshot.TopIPs = append(snapshot.TopIPs, IPCount{
			IP:    ip,
			Count: count,
		})
	}
	sort.Slice(snapshot.TopIPs, func(i, j int) bool {
		return snapshot.TopIPs[i].Count > snapshot.TopIPs[j].Count
	})

	return snapshot
}

// GetTotalLines retorna el total de líneas procesadas.
func (s *Stats) GetTotalLines() int64 {
	return s.totalLines.Load()
}

// GetTotalAttacks retorna el total de ataques detectados.
func (s *Stats) GetTotalAttacks() int64 {
	return s.totalAttacks.Load()
}

// GetAttackCount retorna el conteo de un patrón específico.
func (s *Stats) GetAttackCount(pattern string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.attackCounts[pattern]
}

// GetIPCount retorna el conteo de ataques de una IP específica.
func (s *Stats) GetIPCount(ip string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ipCounts[ip]
}

// GetSeverityCount retorna el conteo de ataques de un nivel de severidad.
func (s *Stats) GetSeverityCount(severity Severity) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if severity >= 0 && int(severity) < len(s.bySeverity) {
		return s.bySeverity[severity]
	}
	return 0
}
