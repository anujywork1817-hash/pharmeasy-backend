package routes

import (
	"fmt"
	"pharmeasy-backend/handlers"
	"pharmeasy-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Serve uploaded files
	r.Static("/uploads", "./uploads")

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ─── User Auth routes (public) ────────────────────────────────────────────
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/forgot-password/send-otp", handlers.SendForgotPasswordOTP)
		auth.POST("/forgot-password/reset", handlers.VerifyOTPAndResetPassword)
	}

	// ─── Doctor Auth routes (PUBLIC - no middleware) ───────────────────────────
	doctorAuth := r.Group("/api/doctor/auth")
	{
		doctorAuth.POST("/register", handlers.DoctorRegister)
		doctorAuth.POST("/login", handlers.DoctorLogin)
	}

	// ─── Public routes ────────────────────────────────────────────────────────
	public := r.Group("/api")
	{
		public.GET("/doctors", handlers.GetDoctors)
		public.GET("/doctors/:id", handlers.GetDoctorByID)
		public.GET("/doctors/:id/booked-slots", handlers.GetBookedSlots)
	}

	// ─── Protected user routes ────────────────────────────────────────────────
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// User
		api.GET("/profile", handlers.GetProfile)

		// Medicines
		api.GET("/medicines", handlers.GetMedicines)
		api.GET("/medicines/:id", handlers.GetMedicineByID)

		// Reviews
		api.POST("/reviews", handlers.AddReview)
		api.GET("/medicines/:id/reviews", handlers.GetReviews)
		api.GET("/medicines/:id/my-review", handlers.GetMyReview)
		api.DELETE("/medicines/:id/review", handlers.DeleteReview)

		// Orders
		api.POST("/orders", handlers.PlaceOrder)
		api.GET("/orders/my", handlers.GetMyOrders)
		api.GET("/orders/:id", handlers.GetOrderByID)
		api.GET("/orders/:id/tracking", handlers.GetOrderTracking)
		api.PUT("/orders/:id/status", handlers.UpdateOrderStatus)
		api.POST("/orders/:id/simulate", handlers.SimulateTracking)
		api.PATCH("/orders/:id/cancel", handlers.CancelOrder)

		// Prescription
		api.POST("/prescription/upload", handlers.UploadPrescription)

		// Payment
		api.POST("/payment/create-order", handlers.CreatePaymentOrder)
		api.POST("/payment/verify", handlers.VerifyPayment)
		api.GET("/payment/history", handlers.GetPaymentHistory)

		// Coupons
		api.POST("/coupons/validate", handlers.ValidateCoupon)
		api.GET("/coupons", handlers.GetActiveCoupons)
		api.POST("/coupons", handlers.CreateCoupon)
		api.PUT("/orders/:id/cancel", handlers.CancelOrder)

		// Appointments
		api.POST("/appointments", handlers.BookAppointment)
		api.GET("/appointments/my", handlers.GetMyAppointments)
		api.PUT("/appointments/:id/cancel", handlers.CancelAppointment)
	}

	// ─── Protected doctor routes ──────────────────────────────────────────────
	doctorAPI := r.Group("/api/doctor")
	doctorAPI.Use(middleware.DoctorAuthMiddleware())
	{
		doctorAPI.GET("/profile", handlers.DoctorGetProfile)
		doctorAPI.PUT("/profile", handlers.DoctorUpdateProfile)
		doctorAPI.GET("/appointments/today", handlers.DoctorGetTodayAppointments)
		doctorAPI.GET("/appointments", handlers.DoctorGetAllAppointments)
		doctorAPI.PUT("/appointments/:id", handlers.DoctorUpdateAppointment)
		doctorAPI.GET("/earnings", handlers.DoctorGetEarnings)
		doctorAPI.PUT("/availability", handlers.DoctorUpdateAvailability)
		doctorAPI.GET("/reviews", handlers.DoctorGetReviews)
	}

	fmt.Println("✅ Routes registered including PATCH cancel")

	return r
}
