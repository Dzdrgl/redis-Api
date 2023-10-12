package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) HandleSimulation(w http.ResponseWriter, r *http.Request) {
	log.Println("Simulation - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	var simInfo models.SimulationInfo
	if err := json.NewDecoder(r.Body).Decode(&simInfo); err != nil {
		log.Printf("Simulation - Failed to decode JSON payload: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}
	for i := 1; i <= simInfo.Usercount; i++ {
		if err := h.createSimUser(); err != nil {
			log.Printf("Simulation - user #%d: %v", i, err)
			errorResponse(w, http.StatusInternalServerError, "Simulation failed during user creation.")
			return
		}
	}
	log.Printf("Successfully created %d simulated user(s).", simInfo.Usercount)
	if err := h.matchSimulation(); err != nil {
		log.Printf("Simulation - %v", err)
		errorResponse(w, http.StatusInternalServerError, "Simulation failed during match operation.")
		return
	}
	result := models.SuccessResponse{
		Status: true,
		Result: "Simulation completed successfully.",
	}
	successResponse(w, result)
	log.Println("Simulation completed successfully.")
}
