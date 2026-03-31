'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { 
  FaChartLine, 
  FaShieldAlt, 
  FaUsers, 
  FaBolt, 
  FaCheckCircle, 
  FaArrowRight,
  FaMoon,
  FaSun,
  FaRocket,
  FaClock,
  FaLock,
  FaChartBar
} from 'react-icons/fa';

export default function Home() {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isDark, setIsDark] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    const savedTheme = localStorage.getItem('theme');
    
    if (savedTheme) {
      setIsDark(savedTheme === 'dark');
    }
    
    if (token) {
      setIsAuthenticated(true);
      setTimeout(() => router.push('/dashboard'), 1500);
    } else {
      setIsLoading(false);
    }
  }, [router]);

  const toggleTheme = () => {
    const newTheme = !isDark;
    setIsDark(newTheme);
    localStorage.setItem('theme', newTheme ? 'dark' : 'light');
  };

  if (isLoading || isAuthenticated) {
    return (
      <div className={`min-h-screen flex items-center justify-center ${isDark ? 'bg-slate-950' : 'bg-gradient-to-br from-blue-50 to-indigo-100'}`}>
        <div className="flex flex-col items-center gap-4">
          <div className={`w-16 h-16 rounded-2xl flex items-center justify-center ${isDark ? 'bg-gradient-to-br from-blue-500 to-purple-600' : 'bg-gradient-to-br from-blue-600 to-indigo-600'} shadow-2xl animate-pulse`}>
            <FaChartLine className="text-white text-2xl" />
          </div>
          <p className={`text-sm font-medium ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
            {isAuthenticated ? 'جاري التوجيه إلى لوحة التحكم...' : 'جاري التحميل...'}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className={`min-h-screen transition-colors duration-300 ${isDark ? 'bg-slate-950' : 'bg-white'}`}>
      {/* Navigation */}
      <nav className={`border-b backdrop-blur-md sticky top-0 z-50 transition-colors ${
        isDark 
          ? 'border-slate-800 bg-slate-950/80' 
          : 'border-slate-200 bg-white/80'
      }`}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-3">
              <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                isDark 
                  ? 'bg-gradient-to-br from-blue-500 to-purple-600' 
                  : 'bg-gradient-to-br from-blue-600 to-indigo-600'
              } shadow-lg`}>
                <FaChartLine className="text-white text-lg" />
              </div>
              <span className={`text-xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>
                CopyTrader
              </span>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={toggleTheme}
                className={`p-2.5 rounded-lg transition-all ${
                  isDark 
                    ? 'bg-slate-800 hover:bg-slate-700 text-yellow-400' 
                    : 'bg-slate-100 hover:bg-slate-200 text-slate-700'
                }`}
                aria-label="Toggle theme"
              >
                {isDark ? <FaSun className="text-lg" /> : <FaMoon className="text-lg" />}
              </button>
              <Link 
                href="/auth/login"
                className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                  isDark 
                    ? 'text-slate-300 hover:text-white hover:bg-slate-800' 
                    : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
                }`}
              >
                تسجيل الدخول
              </Link>
              <Link 
                href="/auth/register"
                className={`px-6 py-2.5 rounded-lg font-medium transition-all shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 ${
                  isDark 
                    ? 'bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white' 
                    : 'bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white'
                }`}
              >
                ابدأ الآن
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-16">
        <div className="text-center">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full mb-6 bg-gradient-to-r from-blue-500/10 to-purple-500/10 border border-blue-500/20">
            <FaRocket className={isDark ? 'text-blue-400' : 'text-blue-600'} />
            <span className={`text-sm font-medium ${isDark ? 'text-blue-400' : 'text-blue-600'}`}>
              منصة التداول الأكثر تطوراً
            </span>
          </div>
          
          <h1 className={`text-5xl md:text-7xl font-bold mb-6 ${isDark ? 'text-white' : 'text-slate-900'}`}>
            انسخ صفقات
            <span className={`block mt-2 bg-gradient-to-r ${
              isDark 
                ? 'from-blue-400 via-purple-400 to-pink-400' 
                : 'from-blue-600 via-purple-600 to-pink-600'
            } bg-clip-text text-transparent`}>
              المحترفين تلقائياً
            </span>
          </h1>
          
          <p className={`text-xl md:text-2xl mb-10 max-w-3xl mx-auto leading-relaxed ${
            isDark ? 'text-slate-400' : 'text-slate-600'
          }`}>
            منصة Copy Trading احترافية تتيح لك نسخ صفقات أفضل المتداولين
            <br />
            مع حماية كاملة وإدارة مخاطر متقدمة على Binance
          </p>
          
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Link 
              href="/auth/register"
              className={`group px-8 py-4 rounded-xl font-semibold text-lg transition-all shadow-xl hover:shadow-2xl transform hover:-translate-y-1 flex items-center gap-3 ${
                isDark 
                  ? 'bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white' 
                  : 'bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white'
              }`}
            >
              ابدأ التداول مجاناً
              <FaArrowRight className="group-hover:translate-x-1 transition-transform" />
            </Link>
            <Link 
              href="/auth/login"
              className={`px-8 py-4 rounded-xl font-semibold text-lg transition-all border-2 ${
                isDark 
                  ? 'border-slate-700 hover:border-slate-600 text-white hover:bg-slate-800' 
                  : 'border-slate-300 hover:border-slate-400 text-slate-900 hover:bg-slate-50'
              }`}
            >
              تسجيل الدخول
            </Link>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-8 mt-16 max-w-3xl mx-auto">
            <div className={`p-6 rounded-2xl ${isDark ? 'bg-slate-900/50' : 'bg-slate-50'}`}>
              <div className={`text-3xl font-bold mb-1 ${
                isDark ? 'text-blue-400' : 'text-blue-600'
              }`}>99.9%</div>
              <div className={`text-sm ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
                وقت التشغيل
              </div>
            </div>
            <div className={`p-6 rounded-2xl ${isDark ? 'bg-slate-900/50' : 'bg-slate-50'}`}>
              <div className={`text-3xl font-bold mb-1 ${
                isDark ? 'text-purple-400' : 'text-purple-600'
              }`}>{"\u003C100ms"}</div>
              <div className={`text-sm ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
                سرعة التنفيذ
              </div>
            </div>
            <div className={`p-6 rounded-2xl ${isDark ? 'bg-slate-900/50' : 'bg-slate-50'}`}>
              <div className={`text-3xl font-bold mb-1 ${
                isDark ? 'text-pink-400' : 'text-pink-600'
              }`}>10K+</div>
              <div className={`text-sm ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
                مستخدم نشط
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className={`py-20 ${isDark ? 'bg-slate-900/30' : 'bg-slate-50'}`}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className={`text-4xl md:text-5xl font-bold mb-4 ${isDark ? 'text-white' : 'text-slate-900'}`}>
              لماذا CopyTrader؟
            </h2>
            <p className={`text-lg ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
              تقنيات متقدمة لتجربة تداول سلسة وآمنة
            </p>
          </div>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            <FeatureCard
              icon={<FaBolt />}
              title="تنفيذ فوري"
              description="نسخ الصفقات بسرعة البرق مع معمارية موجهة بالأحداث"
              isDark={isDark}
              gradient="from-yellow-500 to-orange-500"
            />
            <FeatureCard
              icon={<FaLock />}
              title="أمان متقدم"
              description="تشفير AES-256-GCM لحماية مفاتيح API الخاصة بك"
              isDark={isDark}
              gradient="from-green-500 to-emerald-500"
            />
            <FeatureCard
              icon={<FaUsers />}
              title="متعدد المستخدمين"
              description="كل مستخدم له مفاتيح وحدود تداول مخصصة"
              isDark={isDark}
              gradient="from-blue-500 to-cyan-500"
            />
            <FeatureCard
              icon={<FaChartBar />}
              title="إدارة المخاطر"
              description="Circuit Breaker و Rate Limiting لحماية رأس المال"
              isDark={isDark}
              gradient="from-purple-500 to-pink-500"
            />
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className={`text-4xl md:text-5xl font-bold mb-4 ${isDark ? 'text-white' : 'text-slate-900'}`}>
              كيف يعمل؟
            </h2>
            <p className={`text-lg ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
              ابدأ التداول في 3 خطوات بسيطة
            </p>
          </div>
          
          <div className="grid md:grid-cols-3 gap-8">
            <StepCard
              number="1"
              title="سجل حسابك"
              description="أنشئ حساب مجاني وربطه بـ Binance API بأمان تام"
              isDark={isDark}
            />
            <StepCard
              number="2"
              title="اختر متداول"
              description="تصفح قائمة المتداولين المحترفين واختر الأنسب لك"
              isDark={isDark}
            />
            <StepCard
              number="3"
              title="ابدأ النسخ"
              description="الصفقات تُنسخ تلقائياً إلى حسابك في الوقت الفعلي"
              isDark={isDark}
            />
          </div>
        </div>
      </section>

      {/* Pricing Plans */}
      <section className={`py-20 ${isDark ? 'bg-slate-900/30' : 'bg-slate-50'}`}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className={`text-4xl md:text-5xl font-bold mb-4 ${isDark ? 'text-white' : 'text-slate-900'}`}>
              خطط الاشتراك
            </h2>
            <p className={`text-lg ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
              اختر الخطة المناسبة لاحتياجاتك
            </p>
          </div>
          
          <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
            <PricingCard
              name="Basic"
              price="مجاني"
              features={[
                "حد تعرض: 10%",
                "10 أوامر/دقيقة",
                "دعم أساسي",
                "متداول واحد",
                "تقارير أسبوعية"
              ]}
              isDark={isDark}
            />
            <PricingCard
              name="Pro"
              price="$49"
              period="/شهر"
              features={[
                "حد تعرض: 30%",
                "50 أمر/دقيقة",
                "دعم أولوية",
                "5 متداولين",
                "تحليلات متقدمة",
                "تقارير يومية"
              ]}
              highlighted
              isDark={isDark}
            />
            <PricingCard
              name="Enterprise"
              price="$199"
              period="/شهر"
              features={[
                "حد تعرض: 50%",
                "200 أمر/دقيقة",
                "دعم VIP 24/7",
                "متداولين غير محدودين",
                "API مخصص",
                "تقارير فورية"
              ]}
              isDark={isDark}
            />
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className={`rounded-3xl p-12 text-center shadow-2xl ${
            isDark 
              ? 'bg-gradient-to-br from-blue-600 via-purple-600 to-pink-600' 
              : 'bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500'
          }`}>
            <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
              جاهز للبدء؟
            </h2>
            <p className="text-xl text-white/90 mb-8">
              انضم إلى آلاف المتداولين الذين يحققون أرباحاً مع CopyTrader
            </p>
            <Link 
              href="/auth/register"
              className="inline-flex items-center gap-3 bg-white text-blue-600 px-8 py-4 rounded-xl font-semibold text-lg hover:bg-slate-50 transition-all shadow-xl hover:shadow-2xl transform hover:-translate-y-1"
            >
              ابدأ مجاناً الآن
              <FaArrowRight />
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className={`border-t py-12 ${
        isDark 
          ? 'border-slate-800 bg-slate-950' 
          : 'border-slate-200 bg-white'
      }`}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row justify-between items-center gap-4">
            <div className="flex items-center gap-3">
              <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                isDark 
                  ? 'bg-gradient-to-br from-blue-500 to-purple-600' 
                  : 'bg-gradient-to-br from-blue-600 to-indigo-600'
              }`}>
                <FaChartLine className="text-white text-lg" />
              </div>
              <span className={`text-xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>
                CopyTrader
              </span>
            </div>
            <p className={isDark ? 'text-slate-400' : 'text-slate-600'}>
              &copy; 2026 CopyTrader. جميع الحقوق محفوظة.
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ 
  icon, 
  title, 
  description, 
  isDark,
  gradient 
}: { 
  icon: React.ReactNode; 
  title: string; 
  description: string;
  isDark: boolean;
  gradient: string;
}) {
  return (
    <div className={`group p-8 rounded-2xl transition-all hover:scale-105 ${
      isDark 
        ? 'bg-slate-900/50 hover:bg-slate-800/50 border border-slate-800 hover:border-slate-700' 
        : 'bg-white hover:bg-slate-50 border border-slate-200 hover:border-slate-300 shadow-lg hover:shadow-xl'
    }`}>
      <div className={`w-14 h-14 rounded-xl flex items-center justify-center mb-5 bg-gradient-to-br ${gradient} shadow-lg`}>
        <div className="text-white text-2xl">{icon}</div>
      </div>
      <h3 className={`text-xl font-bold mb-3 ${isDark ? 'text-white' : 'text-slate-900'}`}>
        {title}
      </h3>
      <p className={isDark ? 'text-slate-400' : 'text-slate-600'}>
        {description}
      </p>
    </div>
  );
}

function StepCard({ 
  number, 
  title, 
  description,
  isDark 
}: { 
  number: string; 
  title: string; 
  description: string;
  isDark: boolean;
}) {
  return (
    <div className="relative">
      <div className={`p-8 rounded-2xl h-full ${
        isDark 
          ? 'bg-slate-900/50 border border-slate-800' 
          : 'bg-white border border-slate-200 shadow-lg'
      }`}>
        <div className={`w-16 h-16 rounded-2xl flex items-center justify-center text-2xl font-bold mb-6 ${
          isDark 
            ? 'bg-gradient-to-br from-blue-500 to-purple-600 text-white' 
            : 'bg-gradient-to-br from-blue-600 to-indigo-600 text-white'
        } shadow-lg`}>
          {number}
        </div>
        <h3 className={`text-2xl font-bold mb-3 ${isDark ? 'text-white' : 'text-slate-900'}`}>
          {title}
        </h3>
        <p className={`text-lg ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
          {description}
        </p>
      </div>
    </div>
  );
}

function PricingCard({ 
  name, 
  price, 
  period,
  features, 
  highlighted = false,
  isDark
}: { 
  name: string; 
  price: string;
  period?: string;
  features: string[]; 
  highlighted?: boolean;
  isDark: boolean;
}) {
  return (
    <div className={`relative p-8 rounded-3xl transition-all ${
      highlighted 
        ? `transform scale-105 ${
            isDark 
              ? 'bg-gradient-to-br from-blue-600 to-purple-600 border-2 border-blue-500' 
              : 'bg-gradient-to-br from-blue-600 to-indigo-600 border-2 border-blue-400'
          } shadow-2xl` 
        : isDark 
          ? 'bg-slate-900/50 border border-slate-800 hover:border-slate-700' 
          : 'bg-white border border-slate-200 hover:border-slate-300 shadow-lg hover:shadow-xl'
    }`}>
      {highlighted && (
        <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
          <span className="bg-yellow-400 text-slate-900 px-4 py-1 rounded-full text-sm font-bold">
            الأكثر شعبية
          </span>
        </div>
      )}
      
      <h3 className={`text-2xl font-bold mb-2 ${
        highlighted ? 'text-white' : isDark ? 'text-white' : 'text-slate-900'
      }`}>
        {name}
      </h3>
      
      <div className="mb-6">
        <span className={`text-4xl font-bold ${
          highlighted ? 'text-white' : isDark ? 'text-white' : 'text-slate-900'
        }`}>
          {price}
        </span>
        {period && (
          <span className={highlighted ? 'text-white/80' : isDark ? 'text-slate-400' : 'text-slate-600'}>
            {period}
          </span>
        )}
      </div>
      
      <ul className="space-y-4 mb-8">
        {features.map((feature, index) => (
          <li key={index} className="flex items-start gap-3">
            <FaCheckCircle className={`mt-1 flex-shrink-0 ${
              highlighted ? 'text-white' : isDark ? 'text-blue-400' : 'text-blue-600'
            }`} />
            <span className={
              highlighted ? 'text-white/90' : isDark ? 'text-slate-300' : 'text-slate-700'
            }>
              {feature}
            </span>
          </li>
        ))}
      </ul>
      
      <Link 
        href="/auth/register"
        className={`block w-full text-center py-3.5 rounded-xl font-semibold transition-all ${
          highlighted
            ? 'bg-white text-blue-600 hover:bg-slate-50 shadow-lg'
            : isDark
              ? 'bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white'
              : 'bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white'
        }`}
      >
        اختر الخطة
      </Link>
    </div>
  );
}
