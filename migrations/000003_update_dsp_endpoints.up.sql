UPDATE partners
SET endpoint = 'http://dsp-alpha:9001/bid'
WHERE name = 'DSP Alpha';

UPDATE partners
SET endpoint = 'http://dsp-beta:9002/bid'
WHERE name = 'DSP Beta';

UPDATE partners
SET endpoint = 'http://dsp-gamma:9003/bid'
WHERE name = 'DSP Gamma';