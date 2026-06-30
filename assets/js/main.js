(function ($) {
  "use strict";

  $(window).on("load", function () {
    // Hide Preloader
    $("#preloader").fadeOut(500);

    // Isotope Grid Initialization
    const grid = $(".isotope-grid");
    if (grid.length > 0) {
      const iso = grid.isotope({
        itemSelector: ".isotope-item",
        layoutMode: "fitRows",
        percentPosition: true
      });

      // Filtering logic
      $(".filter-btn").on("click", function () {
        const filterValue = $(this).attr("data-filter");
        iso.isotope({ filter: filterValue });
        
        $(".filter-btn").removeClass("active");
        $(this).addClass("active");
      });
    }
  });

  $(document).ready(function () {
    // WOW.js Initialization
    if (typeof WOW !== 'undefined') {
      new WOW({
        animateClass: "animate__animated",
        offset: 100,
        mobile: true,
        live: true,
      }).init();
    }

    // Theme Management
    const darkToggle = $(".dark-toggle");
    const body = $("body");
    const html = document.documentElement;

    // Check saved theme or system preference
    if (html.classList.contains("dark-theme")) {
      body.addClass("dark-theme");
      updateToggleIcons("dark");
    }

    darkToggle.on("click", function () {
      const isDark = !html.classList.contains("dark-theme");
      html.classList.toggle("dark-theme", isDark);
      body.toggleClass("dark-theme", isDark);
      localStorage.setItem("theme", isDark ? "dark" : "light");
      updateToggleIcons(isDark ? "dark" : "light");
    });

    function updateToggleIcons(theme) {
      if (theme === "dark") {
        darkToggle.find("i").removeClass("fa-moon").addClass("fa-sun");
      } else {
        darkToggle.find("i").removeClass("fa-sun").addClass("fa-moon");
      }
    }

    // Mobile Navigation Toggle
    $(".menu-toggle").on("click", function () {
      $(".mobile-menu-inner").toggleClass("active");
    });

    // Contact Form Handler
    const contactForm = $("#contact-form");
    if (contactForm.length > 0) {
      contactForm.on("submit", function (e) {
        e.preventDefault();
        const submitBtn = $(this).find('button[type="submit"]');
        const originalText = submitBtn.html();
        
        // Efecto visual: Transmitiendo...
        submitBtn.html('Transmitting... <i class="fa-solid fa-spinner fa-spin ms-2"></i>').prop('disabled', true);
        
        setTimeout(function() {
          // Efecto visual: Recibido
          submitBtn.html('Data Received <i class="fa-solid fa-check ms-2"></i>').addClass('btn-success').removeClass('btn-primary');
          
          setTimeout(function() {
            // Realizar el POST nativo de forma programática
            contactForm[0].submit();
          }, 1000); // Espera 1 segundo mostrando el mensaje de éxito antes de redirigir
        }, 1500); // Simula el envío
      });
    }

    // Blog Dynamic Content Loading
    const loadMoreBtn = $("#load-more-trigger");
    const blogPool = $(".blog-load-more-pool");
    const blogFeed = $("#blog-feed");

    if (loadMoreBtn.length > 0 && blogPool.length > 0) {
      loadMoreBtn.on("click", function (e) {
        e.preventDefault();
        
        const btn = $(this);
        const btnText = btn.find("span");
        
        // Update button state
        btn.addClass("loading").prop("disabled", true);
        btnText.text("Loading...");
        
        // Mock processing delay
        setTimeout(function() {
          const items = blogPool.children().detach();
          
          if (items.length > 0) {
            items.each(function(index) {
              const item = $(this);
              item.removeClass("d-none");
              
              // Animated reveal sequence
              setTimeout(function() {
                item.addClass("animate__animated animate__fadeInUp");
                blogFeed.append(item);
              }, index * 150);
            });
            
            // Remove trigger after exhaust
            btn.fadeOut(400, function() {
              $(this).remove();
            });
            
          } else {
            btnText.text("All Articles Loaded");
            btn.removeClass("loading");
          }
        }, 1500);
      });
    }
  });

})(jQuery);
