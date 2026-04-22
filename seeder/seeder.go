package seeder

import (
	"fmt"
	"pharmeasy-backend/models"

	"gorm.io/gorm"
)

func SeedMedicines(db *gorm.DB) {
	var count int64
	db.Model(&models.Medicine{}).Count(&count)
	if count > 0 {
		fmt.Println("✅ Medicines already seeded, skipping.")
		return
	}

	medicines := []models.Medicine{
		// ── Pain Relief ───────────────────────────────────────────────
		{Name: "Paracetamol 500mg", Brand: "Calpol", Description: "Fever and mild pain relief", Price: 45, Discount: 10, Stock: 200, RequiresRx: false, Category: "Pain Relief"},
		{Name: "Ibuprofen 400mg", Brand: "Brufen", Description: "Anti-inflammatory and pain relief", Price: 85, Discount: 15, Stock: 150, RequiresRx: false, Category: "Pain Relief"},
		{Name: "Aspirin 75mg", Brand: "Disprin", Description: "Blood thinner and pain relief", Price: 30, Discount: 5, Stock: 180, RequiresRx: false, Category: "Pain Relief"},
		{Name: "Diclofenac 50mg", Brand: "Voveran", Description: "Joint and muscle pain relief", Price: 65, Discount: 10, Stock: 120, RequiresRx: false, Category: "Pain Relief"},
		{Name: "Nimesulide 100mg", Brand: "Nimulid", Description: "Fever and pain relief", Price: 55, Discount: 8, Stock: 160, RequiresRx: false, Category: "Pain Relief"},
		{Name: "Tramadol 50mg", Brand: "Ultram", Description: "Moderate to severe pain relief", Price: 120, Discount: 5, Stock: 80, RequiresRx: true, Category: "Pain Relief"},
		{Name: "Aceclofenac 100mg", Brand: "Zerodol", Description: "Anti-inflammatory pain relief", Price: 75, Discount: 12, Stock: 130, RequiresRx: false, Category: "Pain Relief"},
		{Name: "Mefenamic Acid 500mg", Brand: "Meftal", Description: "Period and mild pain relief", Price: 50, Discount: 10, Stock: 140, RequiresRx: false, Category: "Pain Relief"},

		// ── Antibiotics ───────────────────────────────────────────────
		{Name: "Amoxicillin 500mg", Brand: "Mox", Description: "Broad spectrum antibiotic", Price: 110, Discount: 5, Stock: 100, RequiresRx: true, Category: "Antibiotic"},
		{Name: "Azithromycin 500mg", Brand: "Zithromax", Description: "Respiratory tract infections", Price: 145, Discount: 10, Stock: 90, RequiresRx: true, Category: "Antibiotic"},
		{Name: "Ciprofloxacin 500mg", Brand: "Ciplox", Description: "Urinary and bacterial infections", Price: 95, Discount: 8, Stock: 110, RequiresRx: true, Category: "Antibiotic"},
		{Name: "Doxycycline 100mg", Brand: "Doxt", Description: "Bacterial and acne treatment", Price: 130, Discount: 10, Stock: 85, RequiresRx: true, Category: "Antibiotic"},
		{Name: "Metronidazole 400mg", Brand: "Flagyl", Description: "Anaerobic bacterial infections", Price: 40, Discount: 5, Stock: 160, RequiresRx: true, Category: "Antibiotic"},
		{Name: "Cetirizine 10mg", Brand: "Zyrtec", Description: "Antihistamine for allergy relief", Price: 35, Discount: 5, Stock: 200, RequiresRx: false, Category: "Antibiotic"},
		{Name: "Levofloxacin 500mg", Brand: "Levaquin", Description: "Wide spectrum antibiotic", Price: 175, Discount: 12, Stock: 70, RequiresRx: true, Category: "Antibiotic"},
		{Name: "Cefixime 200mg", Brand: "Taxim-O", Description: "Third generation cephalosporin", Price: 160, Discount: 8, Stock: 80, RequiresRx: true, Category: "Antibiotic"},

		// ── Vitamins ──────────────────────────────────────────────────
		{Name: "Vitamin C 500mg", Brand: "Limcee", Description: "Immunity booster", Price: 60, Discount: 10, Stock: 250, RequiresRx: false, Category: "Vitamins"},
		{Name: "Vitamin D3 60000 IU", Brand: "Calcirol", Description: "Bone health supplement", Price: 90, Discount: 5, Stock: 200, RequiresRx: false, Category: "Vitamins"},
		{Name: "Vitamin B12 500mcg", Brand: "Neurobion", Description: "Nerve and energy support", Price: 115, Discount: 8, Stock: 180, RequiresRx: false, Category: "Vitamins"},
		{Name: "Multivitamin Daily", Brand: "Supradyn", Description: "Complete daily nutrition", Price: 250, Discount: 15, Stock: 150, RequiresRx: false, Category: "Vitamins"},
		{Name: "Zinc 50mg", Brand: "Zincovit", Description: "Immunity and skin health", Price: 80, Discount: 10, Stock: 170, RequiresRx: false, Category: "Vitamins"},
		{Name: "Iron + Folic Acid", Brand: "Feronia", Description: "Anaemia and pregnancy support", Price: 70, Discount: 5, Stock: 160, RequiresRx: false, Category: "Vitamins"},
		{Name: "Omega-3 Fish Oil", Brand: "Healthviva", Description: "Heart and brain health", Price: 320, Discount: 20, Stock: 120, RequiresRx: false, Category: "Vitamins"},
		{Name: "Calcium + Vitamin D", Brand: "Shelcal", Description: "Bone strength supplement", Price: 135, Discount: 10, Stock: 140, RequiresRx: false, Category: "Vitamins"},

		// ── Diabetes ──────────────────────────────────────────────────
		{Name: "Metformin 500mg", Brand: "Glycomet", Description: "Type 2 diabetes management", Price: 55, Discount: 5, Stock: 200, RequiresRx: true, Category: "Diabetes"},
		{Name: "Glimepiride 2mg", Brand: "Amaryl", Description: "Blood sugar control", Price: 85, Discount: 8, Stock: 130, RequiresRx: true, Category: "Diabetes"},
		{Name: "Sitagliptin 100mg", Brand: "Januvia", Description: "Type 2 diabetes oral medicine", Price: 520, Discount: 5, Stock: 60, RequiresRx: true, Category: "Diabetes"},
		{Name: "Dapagliflozin 10mg", Brand: "Forxiga", Description: "SGLT2 inhibitor for diabetes", Price: 680, Discount: 10, Stock: 50, RequiresRx: true, Category: "Diabetes"},
		{Name: "Insulin Glargine", Brand: "Lantus", Description: "Long acting insulin", Price: 950, Discount: 5, Stock: 40, RequiresRx: true, Category: "Diabetes"},
		{Name: "Voglibose 0.3mg", Brand: "Volix", Description: "Post meal sugar control", Price: 95, Discount: 8, Stock: 110, RequiresRx: true, Category: "Diabetes"},

		// ── Heart & BP ────────────────────────────────────────────────
		{Name: "Amlodipine 5mg", Brand: "Norvasc", Description: "Blood pressure management", Price: 65, Discount: 8, Stock: 180, RequiresRx: true, Category: "Heart & BP"},
		{Name: "Atorvastatin 10mg", Brand: "Lipitor", Description: "Cholesterol lowering", Price: 110, Discount: 10, Stock: 150, RequiresRx: true, Category: "Heart & BP"},
		{Name: "Losartan 50mg", Brand: "Cozaar", Description: "Hypertension management", Price: 90, Discount: 5, Stock: 130, RequiresRx: true, Category: "Heart & BP"},
		{Name: "Telmisartan 40mg", Brand: "Telma", Description: "ARB for blood pressure", Price: 120, Discount: 10, Stock: 120, RequiresRx: true, Category: "Heart & BP"},
		{Name: "Metoprolol 25mg", Brand: "Betaloc", Description: "Beta blocker for heart rate", Price: 75, Discount: 8, Stock: 140, RequiresRx: true, Category: "Heart & BP"},
		{Name: "Rosuvastatin 10mg", Brand: "Crestor", Description: "Statin for cholesterol", Price: 145, Discount: 12, Stock: 110, RequiresRx: true, Category: "Heart & BP"},

		// ── Stomach & Digestion ───────────────────────────────────────
		{Name: "Omeprazole 20mg", Brand: "Prilosec", Description: "Acidity and ulcer treatment", Price: 55, Discount: 10, Stock: 200, RequiresRx: false, Category: "Stomach"},
		{Name: "Pantoprazole 40mg", Brand: "Pantocid", Description: "Gastric acid reducer", Price: 70, Discount: 8, Stock: 180, RequiresRx: false, Category: "Stomach"},
		{Name: "Domperidone 10mg", Brand: "Motilium", Description: "Nausea and vomiting relief", Price: 45, Discount: 5, Stock: 170, RequiresRx: false, Category: "Stomach"},
		{Name: "Ondansetron 4mg", Brand: "Zofran", Description: "Anti-nausea medication", Price: 80, Discount: 10, Stock: 130, RequiresRx: false, Category: "Stomach"},
		{Name: "Loperamide 2mg", Brand: "Imodium", Description: "Diarrhoea relief", Price: 40, Discount: 5, Stock: 160, RequiresRx: false, Category: "Stomach"},
		{Name: "Lactulose Syrup", Brand: "Duphalac", Description: "Constipation relief", Price: 130, Discount: 8, Stock: 100, RequiresRx: false, Category: "Stomach"},
		{Name: "ORS Sachet", Brand: "Electral", Description: "Rehydration salts", Price: 25, Discount: 0, Stock: 300, RequiresRx: false, Category: "Stomach"},

		// ── Skin Care ─────────────────────────────────────────────────
		{Name: "Clotrimazole Cream", Brand: "Candid", Description: "Antifungal skin cream", Price: 65, Discount: 5, Stock: 150, RequiresRx: false, Category: "Skin Care"},
		{Name: "Betamethasone Cream", Brand: "Betnovate", Description: "Skin inflammation relief", Price: 75, Discount: 8, Stock: 120, RequiresRx: true, Category: "Skin Care"},
		{Name: "Mupirocin Ointment", Brand: "Bactroban", Description: "Bacterial skin infection", Price: 110, Discount: 5, Stock: 90, RequiresRx: true, Category: "Skin Care"},
		{Name: "Calamine Lotion", Brand: "Lacto Calamine", Description: "Itching and rash relief", Price: 85, Discount: 10, Stock: 140, RequiresRx: false, Category: "Skin Care"},
		{Name: "Salicylic Acid 6%", Brand: "Saslic", Description: "Acne and dandruff treatment", Price: 95, Discount: 12, Stock: 110, RequiresRx: false, Category: "Skin Care"},

		// ── Respiratory ───────────────────────────────────────────────
		{Name: "Salbutamol 100mcg", Brand: "Ventolin", Description: "Asthma reliever inhaler", Price: 180, Discount: 5, Stock: 80, RequiresRx: true, Category: "Respiratory"},
		{Name: "Montelukast 10mg", Brand: "Singulair", Description: "Asthma and allergy control", Price: 160, Discount: 10, Stock: 90, RequiresRx: true, Category: "Respiratory"},
		{Name: "Budesonide Inhaler", Brand: "Pulmicort", Description: "Asthma preventer inhaler", Price: 420, Discount: 8, Stock: 50, RequiresRx: true, Category: "Respiratory"},
		{Name: "Levosalbutamol 1mg", Brand: "Levolin", Description: "Bronchospasm relief", Price: 95, Discount: 5, Stock: 100, RequiresRx: true, Category: "Respiratory"},
		{Name: "Ambroxol 30mg", Brand: "Mucosolvan", Description: "Cough and mucus relief", Price: 60, Discount: 8, Stock: 150, RequiresRx: false, Category: "Respiratory"},
		{Name: "Dextromethorphan Syrup", Brand: "Benylin", Description: "Dry cough suppressant", Price: 85, Discount: 5, Stock: 120, RequiresRx: false, Category: "Respiratory"},
	}

	result := db.Create(&medicines)
	if result.Error != nil {
		fmt.Println("❌ Seeding failed:", result.Error)
		return
	}
	fmt.Printf("✅ Seeded %d medicines successfully!\n", len(medicines))
}
