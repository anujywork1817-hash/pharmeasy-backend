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

	// Auth routes (public)
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/forgot-password/send-otp", handlers.SendForgotPasswordOTP) // ✅ add
		auth.POST("/forgot-password/reset", handlers.VerifyOTPAndResetPassword)

	}

	// Public routes
	public := r.Group("/api")
	{
		public.GET("/doctors", handlers.GetDoctors)
		public.GET("/doctors/:id", handlers.GetDoctorByID)
		public.GET("/doctors/:id/booked-slots", handlers.GetBookedSlots)
	}
	// Protected routes
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

		// Inside protected api group:
		api.POST("/coupons/validate", handlers.ValidateCoupon)
		api.GET("/coupons", handlers.GetActiveCoupons)
		api.POST("/coupons", handlers.CreateCoupon) // admin
		api.PUT("/orders/:id/cancel", handlers.CancelOrder)

		api.POST("/appointments", handlers.BookAppointment)
		api.GET("/appointments/my", handlers.GetMyAppointments)
		api.PUT("/appointments/:id/cancel", handlers.CancelAppointment)
	}

	fmt.Println("✅ Routes registered including PATCH cancel")

	return r
}
