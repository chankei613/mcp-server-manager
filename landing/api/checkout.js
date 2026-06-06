const Stripe = require('stripe');

const SITE_URL = process.env.SITE_URL || 'https://mcpilot.app';

module.exports = async (req, res) => {
  const stripeSecretKey = process.env.STRIPE_SECRET_KEY;
  const priceId = process.env.STRIPE_PRICE_ID;

  if (!stripeSecretKey || !priceId) {
    res.status(500).send('Server configuration error');
    return;
  }

  const stripe = Stripe(stripeSecretKey);

  try {
    const session = await stripe.checkout.sessions.create({
      mode: 'payment',
      line_items: [{ price: priceId, quantity: 1 }],
      success_url: `${SITE_URL}/thank-you.html?session_id={CHECKOUT_SESSION_ID}`,
      cancel_url: `${SITE_URL}/`,
    });

    res.redirect(303, session.url);
  } catch (err) {
    console.error('Checkout session error:', err.message);
    res.status(500).send('Failed to create checkout session');
  }
};
