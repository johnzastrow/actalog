// ========================================
// ATHLETIC BRUTALISM INTERACTIONS
// ActaLog - Elite CrossFit Performance Tracker
// ========================================

document.addEventListener('DOMContentLoaded', () => {
    // Counter Animation
    animateCounters();

    // Intersection Observer for scroll animations
    setupScrollAnimations();

    // Smooth scroll for anchor links
    setupSmoothScroll();
});

// ========================================
// COUNTER ANIMATIONS
// ========================================
function animateCounters() {
    const counters = document.querySelectorAll('[data-target]');

    const animateValue = (element, start, end, duration, isDecimal = false) => {
        const startTime = performance.now();

        const step = (currentTime) => {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);

            // Easing function (ease-out)
            const easeOut = 1 - Math.pow(1 - progress, 3);

            const current = start + (end - start) * easeOut;

            if (isDecimal) {
                element.textContent = current.toFixed(2);
            } else {
                element.textContent = Math.floor(current);
            }

            if (progress < 1) {
                requestAnimationFrame(step);
            } else {
                if (isDecimal) {
                    element.textContent = end.toFixed(2);
                } else {
                    element.textContent = end;
                }
            }
        };

        requestAnimationFrame(step);
    };

    // Trigger counters when they come into view
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting && !entry.target.dataset.animated) {
                const target = parseFloat(entry.target.dataset.target);
                const isDecimal = entry.target.dataset.target.includes('.');
                const duration = 2000; // 2 seconds

                animateValue(entry.target, 0, target, duration, isDecimal);
                entry.target.dataset.animated = 'true';
            }
        });
    }, {
        threshold: 0.5
    });

    counters.forEach(counter => observer.observe(counter));
}

// ========================================
// SCROLL ANIMATIONS
// ========================================
function setupScrollAnimations() {
    const elements = document.querySelectorAll('.feature-card, .tech-item, .stat-item');

    const observer = new IntersectionObserver((entries) => {
        entries.forEach((entry, index) => {
            if (entry.isIntersecting) {
                // Stagger the animation
                setTimeout(() => {
                    entry.target.style.opacity = '1';
                    entry.target.style.transform = 'translateY(0)';
                }, index * 100);
            }
        });
    }, {
        threshold: 0.1
    });

    elements.forEach(element => {
        // Set initial state
        element.style.opacity = '0';
        element.style.transform = 'translateY(30px)';
        element.style.transition = 'opacity 0.6s ease, transform 0.6s ease';

        observer.observe(element);
    });
}

// ========================================
// SMOOTH SCROLL
// ========================================
function setupSmoothScroll() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            const href = this.getAttribute('href');

            // Skip empty anchors
            if (href === '#') return;

            e.preventDefault();

            const target = document.querySelector(href);
            if (target) {
                const offsetTop = target.offsetTop - 80; // Account for fixed nav

                window.scrollTo({
                    top: offsetTop,
                    behavior: 'smooth'
                });
            }
        });
    });
}

// ========================================
// BUTTON RIPPLE EFFECT
// ========================================
document.querySelectorAll('.btn').forEach(button => {
    button.addEventListener('click', function(e) {
        const rect = this.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        const ripple = document.createElement('span');
        ripple.className = 'ripple';
        ripple.style.left = x + 'px';
        ripple.style.top = y + 'px';

        this.appendChild(ripple);

        setTimeout(() => ripple.remove(), 600);
    });
});

// ========================================
// PARALLAX EFFECT ON HERO
// ========================================
let lastScrollY = window.scrollY;

window.addEventListener('scroll', () => {
    const scrollY = window.scrollY;
    const mockup = document.querySelector('.mockup-container');

    if (mockup) {
        // Parallax effect: move slower than scroll
        const parallaxSpeed = 0.5;
        const translateY = scrollY * parallaxSpeed;

        mockup.style.transform = `translateY(${translateY}px)`;
    }

    lastScrollY = scrollY;
}, { passive: true });

// ========================================
// NAVIGATION BACKGROUND ON SCROLL
// ========================================
window.addEventListener('scroll', () => {
    const nav = document.querySelector('.nav');

    if (window.scrollY > 50) {
        nav.style.background = 'rgba(10, 22, 40, 0.98)';
    } else {
        nav.style.background = 'rgba(10, 22, 40, 0.95)';
    }
}, { passive: true });

// ========================================
// CODE BLOCK COPY FUNCTIONALITY (if needed)
// ========================================
document.querySelectorAll('.code-block').forEach(block => {
    const copyButton = document.createElement('button');
    copyButton.className = 'code-copy-btn';
    copyButton.textContent = 'COPY';
    copyButton.style.cssText = `
        position: absolute;
        top: 1rem;
        right: 1rem;
        padding: 0.5rem 1rem;
        background: var(--color-primary);
        color: var(--color-bg-dark);
        border: none;
        border-radius: 4px;
        font-family: var(--font-mono);
        font-size: 0.75rem;
        cursor: pointer;
        opacity: 0;
        transition: opacity 0.3s ease;
    `;

    block.style.position = 'relative';
    block.appendChild(copyButton);

    block.addEventListener('mouseenter', () => {
        copyButton.style.opacity = '1';
    });

    block.addEventListener('mouseleave', () => {
        copyButton.style.opacity = '0';
    });

    copyButton.addEventListener('click', async () => {
        const code = block.querySelector('code').textContent;

        try {
            await navigator.clipboard.writeText(code);
            copyButton.textContent = 'COPIED!';
            setTimeout(() => {
                copyButton.textContent = 'COPY';
            }, 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    });
});

// ========================================
// PERFORMANCE OPTIMIZATION
// ========================================

// Lazy load images when they come into view
const lazyImages = document.querySelectorAll('img[data-src]');

const imageObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            const img = entry.target;
            img.src = img.dataset.src;
            img.removeAttribute('data-src');
            imageObserver.unobserve(img);
        }
    });
});

lazyImages.forEach(img => imageObserver.observe(img));

// ========================================
// MOBILE MENU TOGGLE (if needed)
// ========================================
const createMobileMenu = () => {
    // Only create mobile menu if screen is small
    if (window.innerWidth > 768) return;

    const nav = document.querySelector('.nav-links');
    const container = document.querySelector('.nav-container');

    // Create hamburger button
    const hamburger = document.createElement('button');
    hamburger.className = 'hamburger';
    hamburger.innerHTML = `
        <span></span>
        <span></span>
        <span></span>
    `;
    hamburger.style.cssText = `
        display: flex;
        flex-direction: column;
        gap: 4px;
        background: none;
        border: none;
        cursor: pointer;
        padding: 0.5rem;
    `;

    hamburger.querySelectorAll('span').forEach(span => {
        span.style.cssText = `
            width: 24px;
            height: 2px;
            background: var(--color-primary);
            transition: all 0.3s ease;
        `;
    });

    container.appendChild(hamburger);

    // Toggle menu
    hamburger.addEventListener('click', () => {
        nav.classList.toggle('active');

        if (nav.classList.contains('active')) {
            nav.style.cssText = `
                position: absolute;
                top: 100%;
                left: 0;
                width: 100%;
                background: var(--color-bg-darker);
                flex-direction: column;
                padding: 2rem;
                border-top: 1px solid var(--color-border);
            `;
        } else {
            nav.style.cssText = '';
        }
    });
};

// Initialize mobile menu if needed
if (window.innerWidth <= 768) {
    createMobileMenu();
}

// Re-initialize on resize
let resizeTimer;
window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => {
        if (window.innerWidth <= 768) {
            createMobileMenu();
        }
    }, 250);
});

console.log('🏋️ ActaLog site loaded - Athletic Brutalism activated');
