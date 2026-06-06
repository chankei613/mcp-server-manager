const Stripe = require('stripe');

module.exports = async (req, res) => {
  const { session_id } = req.query;

  if (!session_id) {
    res.status(400).send('Missing session_id');
    return;
  }

  const stripeSecretKey = process.env.STRIPE_SECRET_KEY;
  if (!stripeSecretKey) {
    res.status(500).send('Server configuration error');
    return;
  }

  const stripe = Stripe(stripeSecretKey);

  try {
    const session = await stripe.checkout.sessions.retrieve(session_id);

    if (session.payment_status !== 'paid') {
      res.status(403).send('Payment not completed');
      return;
    }

    const downloadUrl = process.env.MCPILOT_DOWNLOAD_URL;
    if (!downloadUrl) {
      res.status(500).send('Download URL not configured');
      return;
    }

    res.setHeader('Cache-Control', 'no-store');
    res.redirect(302, downloadUrl);
  } catch (err) {
    console.error('Stripe verification error:', err.message);
    res.status(400).send('Invalid or expired session');
  }
};
